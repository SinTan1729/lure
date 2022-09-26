package main

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/rpm"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/cpu"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"
	"go.arsenm.dev/lure/distro"
	"go.arsenm.dev/lure/download"
	"go.arsenm.dev/lure/internal/shutils/decoder"
	"go.arsenm.dev/lure/manager"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// BuildVars represents the script variables required
// to build a package
type BuildVars struct {
	Name          string   `sh:"name,required"`
	Version       string   `sh:"version,required"`
	Release       int      `sh:"release,required"`
	Epoch         uint     `sh:"epoch"`
	Description   string   `sh:"desc"`
	Homepage      string   `sh:"homepage"`
	Architectures []string `sh:"architectures"`
	Licenses      []string `sh:"license"`
	Provides      []string `sh:"provides"`
	Conflicts     []string `sh:"conflicts"`
	Depends       []string `sh:"deps"`
	BuildDepends  []string `sh:"build_deps"`
	Replaces      []string `sh:"replaces"`
	Sources       []string `sh:"sources"`
	Checksums     []string `sh:"checksums"`
	Backup        []string `sh:"backup"`
}

func buildCmd(c *cli.Context) error {
	script := c.String("script")

	mgr := manager.Detect()
	if mgr == nil {
		log.Fatal("Unable to detect supported package manager on system").Send()
	}

	_, pkgNames, err := buildPackage(c.Context, script, mgr)
	if err != nil {
		log.Fatal("Error building package").Err(err).Send()
	}

	log.Info("Package(s) built successfully").Any("names", pkgNames).Send()

	return nil
}

func buildPackage(ctx context.Context, script string, mgr manager.Manager) ([]string, []string, error) {
	info, err := distro.ParseOSRelease(ctx)
	if err != nil {
		return nil, nil, err
	}

	fl, err := os.Open(script)
	if err != nil {
		return nil, nil, err
	}

	file, err := syntax.NewParser().Parse(fl, "lure.sh")
	if err != nil {
		return nil, nil, err
	}

	fl.Close()

	env := genBuildEnv(info)

	runner, err := interp.New(
		interp.Env(expand.ListEnviron(env...)),
		interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
	)
	if err != nil {
		return nil, nil, err
	}

	err = runner.Run(ctx, file)
	if err != nil {
		return nil, nil, err
	}

	dec := decoder.New(info, runner)

	var vars BuildVars
	err = dec.DecodeVars(&vars)
	if err != nil {
		return nil, nil, err
	}

	log.Info("Building package").Str("name", vars.Name).Str("version", vars.Version).Send()

	baseDir := filepath.Join(cacheDir, "pkgs", vars.Name)
	srcdir := filepath.Join(baseDir, "src")
	pkgdir := filepath.Join(baseDir, "pkg")

	err = os.RemoveAll(baseDir)
	if err != nil {
		return nil, nil, err
	}

	err = os.MkdirAll(srcdir, 0o755)
	if err != nil {
		return nil, nil, err
	}

	err = os.MkdirAll(pkgdir, 0o755)
	if err != nil {
		return nil, nil, err
	}

	if len(vars.BuildDepends) > 0 {
		log.Info("Installing build dependencies").Send()
		installPkgs(ctx, vars.BuildDepends, mgr)
	}

	var builtDeps, builtNames, repoDeps []string
	if len(vars.Depends) > 0 {
		log.Info("Installing dependencies").Send()

		scripts, notFound := findPkgs(vars.Depends)
		for _, script := range scripts {
			pkgPaths, pkgNames, err := buildPackage(ctx, script, mgr)
			if err != nil {
				return nil, nil, err
			}
			builtDeps = append(builtDeps, pkgPaths...)
			builtNames = append(builtNames, pkgNames...)
			builtNames = append(builtNames, filepath.Base(filepath.Dir(script)))
		}
		repoDeps = notFound
	}

	log.Info("Downloading sources").Send()

	err = getSources(ctx, srcdir, &vars)
	if err != nil {
		return nil, nil, err
	}

	err = setDirVars(ctx, runner, srcdir, pkgdir)
	if err != nil {
		return nil, nil, err
	}

	fn, ok := dec.GetFunc("build")
	if ok {
		log.Info("Executing build()").Send()

		err = fn(ctx, srcdir)
		if err != nil {
			return nil, nil, err
		}
	}

	fn, ok = dec.GetFunc("package")
	if ok {
		log.Info("Executing package()").Send()

		err = fn(ctx, srcdir)
		if err != nil {
			return nil, nil, err
		}
	}

	uniq(
		&repoDeps,
		&builtDeps,
		&builtNames,
	)

	pkgInfo := &nfpm.Info{
		Name:        vars.Name,
		Description: vars.Description,
		Arch:        runtime.GOARCH,
		Version:     vars.Version,
		Release:     strconv.Itoa(vars.Release),
		Homepage:    vars.Homepage,
		License:     strings.Join(vars.Licenses, ", "),
		Overridables: nfpm.Overridables{
			Depends: append(repoDeps, builtNames...),
		},
	}

	if pkgInfo.Arch == "arm" {
		pkgInfo.Arch = checkARMVariant()
	}

	contents := []*files.Content{}
	filepath.Walk(pkgdir, func(path string, fi os.FileInfo, err error) error {
		trimmed := strings.TrimPrefix(path, pkgdir)

		if fi.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}

			_, err = f.Readdirnames(1)
			if err != io.EOF {
				return nil
			}

			contents = append(contents, &files.Content{
				Source:      path,
				Destination: trimmed,
				Type:        "dir",
				FileInfo: &files.ContentFileInfo{
					MTime: fi.ModTime(),
				},
			})

			f.Close()
			return nil
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}

			contents = append(contents, &files.Content{
				Source:      link,
				Destination: trimmed,
				Type:        "symlink",
				FileInfo: &files.ContentFileInfo{
					MTime: fi.ModTime(),
					Mode:  fi.Mode(),
				},
			})

			return nil
		}

		fileContent := &files.Content{
			Source:      path,
			Destination: trimmed,
			FileInfo: &files.ContentFileInfo{
				MTime: fi.ModTime(),
				Mode:  fi.Mode(),
				Size:  fi.Size(),
			},
		}

		if slices.Contains(vars.Backup, trimmed) {
			fileContent.Type = "config|noreplace"
		}

		contents = append(contents, fileContent)

		return nil
	})

	pkgInfo.Overridables.Contents = contents

	packager, err := nfpm.Get(mgr.Format())
	if err != nil {
		return nil, nil, err
	}

	pkgName := packager.ConventionalFileName(pkgInfo)
	pkgPath := filepath.Join(baseDir, pkgName)

	pkgPaths := append(builtDeps, pkgPath)
	pkgNames := append(builtNames, vars.Name)

	pkgFile, err := os.Create(pkgPath)
	if err != nil {
		return nil, nil, err
	}

	err = packager.Package(pkgInfo, pkgFile)
	if err != nil {
		return nil, nil, err
	}

	uniq(&pkgPaths, &pkgNames)

	return pkgPaths, pkgNames, nil
}

func genBuildEnv(info *distro.OSRelease) []string {
	env := os.Environ()
	env = append(
		env,
		"DISTRO_NAME="+info.Name,
		"DISTRO_PRETTY_NAME="+info.PrettyName,
		"DISTRO_ID="+info.ID,
		"DISTRO_BUILD_ID="+info.BuildID,

		"ARCH="+runtime.GOARCH,
		"NCPU="+strconv.Itoa(runtime.NumCPU()),
	)

	return env
}

func getSources(ctx context.Context, srcdir string, bv *BuildVars) error {
	if len(bv.Sources) != len(bv.Checksums) {
		log.Fatal("The checksums array must be the same length as sources")
	}

	for i, src := range bv.Sources {
		opts := download.GetOptions{
			SourceURL:   src,
			Destination: srcdir,
			EncloseGit:  true,
		}

		if bv.Checksums[i] != "SKIP" {
			checksum, err := hex.DecodeString(bv.Checksums[i])
			if err != nil {
				return err
			}
			opts.SHA256Sum = checksum
		}

		err := download.Get(ctx, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

// setDirVars sets srcdir and pkgdir. It's a very hacky way of doing so,
// but setting the runner's Env and Vars fields doesn't seem to work.
func setDirVars(ctx context.Context, runner *interp.Runner, srcdir, pkgdir string) error {
	cmd := "srcdir='" + srcdir + "'\npkgdir='" + pkgdir + "'\n"
	fl, err := syntax.NewParser().Parse(strings.NewReader(cmd), "vars")
	if err != nil {
		return err
	}
	return runner.Run(ctx, fl)
}

// checkARMVariant checks which variant of ARM lure is running
// on, by using the same detection method as Go itself
func checkARMVariant() string {
	armEnv := os.Getenv("LURE_ARM_VARIANT")
	// ensure value has "arm" prefix, such as arm5 or arm6
	if strings.HasPrefix(armEnv, "arm") {
		return armEnv
	}

	if cpu.ARM.HasVFPv3 {
		return "arm7"
	} else if cpu.ARM.HasVFP {
		return "arm6"
	} else {
		return "arm5"
	}
}

func uniq(ss ...*[]string) {
	for _, s := range ss {
		slices.Sort(*s)
		*s = slices.Compact(*s)
	}
}