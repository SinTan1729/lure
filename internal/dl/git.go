package dl

import (
	"errors"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type GitDownloader struct{}

func (GitDownloader) Name() string {
	return "git"
}

func (GitDownloader) Type() Type {
	return TypeDir
}

func (GitDownloader) MatchURL(u string) bool {
	return strings.HasPrefix(u, "git+")
}

func (GitDownloader) Download(opts Options) (Type, string, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return 0, "", err
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	query := u.Query()

	rev := query.Get("~rev")
	query.Del("~rev")

	depthStr := query.Get("~depth")
	query.Del("~depth")

	recursive := query.Get("~recursive")
	query.Del("~recursive")

	u.RawQuery = query.Encode()

	depth := 0
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			return 0, "", err
		}
	}

	co := &git.CloneOptions{
		URL:               u.String(),
		Depth:             depth,
		Progress:          opts.Progress,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}

	if recursive == "true" {
		co.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	r, err := git.PlainClone(opts.Destination, false, co)
	if err != nil {
		return 0, "", err
	}

	if rev != "" {
		h, err := r.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return 0, "", err
		}

		w, err := r.Worktree()
		if err != nil {
			return 0, "", err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: *h,
		})
		if err != nil {
			return 0, "", err
		}
	}

	name := strings.TrimSuffix(path.Base(u.Path), ".git")
	return TypeDir, name, nil
}

func (GitDownloader) Update(opts Options) (bool, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return false, err
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	query := u.Query()
	query.Del("~rev")

	depthStr := query.Get("~depth")
	query.Del("~depth")

	recursive := query.Get("~recursive")
	query.Del("~recursive")

	u.RawQuery = query.Encode()

	r, err := git.PlainOpen(opts.Destination)
	if err != nil {
		return false, err
	}

	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	depth := 0
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			return false, err
		}
	}

	po := &git.PullOptions{
		Depth:             depth,
		Progress:          opts.Progress,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}

	if recursive == "true" {
		po.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	err = w.Pull(po)
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}