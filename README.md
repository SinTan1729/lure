<img src="assets/logo.png" alt="LURE Logo" width="200">

# LURE (Linux User REpository)

[![Go Report Card](https://goreportcard.com/badge/github.com/sintan1729/lure)](https://goreportcard.com/report/github.com/sintan1729/lure)

LURE is a distro-agnostic build system for Linux, similar to the [AUR](https://wiki.archlinux.org/title/Arch_User_Repository). It is currently in **beta**. Most major bugs have been fixed, and most major features have been added. LURE is ready for general use, but may still break or change occasionally.

LURE is written in pure Go and has zero dependencies after building. The only things LURE requires are a command for privilege elevation such as `sudo`, `doas`, etc. as well as a supported package manager. Currently, LURE supports `apt`, `pacman`, `apk`, `dnf`, `yum`, and `zypper`. If a supported package manager exists on your system, it will be detected and used automatically.

---

## Notice for the fork

There hasn't been any new development in the source repo for a few years. Even though the project remains usable (by its very nature, it
doesn't need constant updates), over time the installation binaries have vanished, and potential issues are being ignored.
For now, I just rebuilt and binaries, and updated installation instructions so that they actually work. I might work on fixing issues if I
have the time, but I have no such plans for now.

If you already have `lure` installed, and would like to switch to this fork, it's recommended that you clear up old artifacts. Just run the
following commands.

```bash
rm -f ~/.config/lure/lure.toml
rm -rf ~/.cache/lure

```

---

## Installation

### Install script

The LURE install script will automatically download and install the appropriate LURE package on your system. To use it, simply run the following command:

```bash
curl -fsSL https://raw.githubusercontent.com/SinTan1729/lure/refs/heads/master/scripts/bootstrap.sh | bash
```

**IMPORTANT**: This will download and run the script from GitHub. Please look through any script you download from the internet (including this one) before running it.

### Packages

Distro packages and binary archives are provided at the latest Gitea release: https://github.com/Sintan1729/lure/releases/latest

LURE is also available on the AUR as [linux-user-repository-bin](https://aur.archlinux.org/packages/linux-user-repository-bin)

### Building from source

To build LURE from source, you'll need Go 1.18 or newer. Once Go is installed, clone this repo and run:

```shell
sudo make install
```

---

## Why?

LURE was created because packaging software for multiple Linux distros can be difficult and error-prone, and installing those packages can be a nightmare for users unless they're available in their distro's official repositories. It automates the process of building and installing unofficial packages.

---

## Documentation

The documentation for LURE is in the [docs](docs) directory in this repo.

---

## Web Interface

LURE has an open source web interface, licensed under the AGPLv3 (https://gitea.elara.ws/lure/lure-web), and it's available at https://lure.sh/.

---

## Repositories

LURE's repos are git repositories that contain a directory for each package, with a `lure.sh` file inside. The `lure.sh` file tells LURE how to build the package and information about it. `lure.sh` scripts are similar to the AUR's PKGBUILD scripts.

---

## Acknowledgements

Thanks to the following projects for making LURE possible:

- https://github.com/mvdan/sh
- https://github.com/go-git/go-git
- https://github.com/mholt/archiver
- https://github.com/goreleaser/nfpm
- https://github.com/charmbracelet/bubbletea
- https://gitlab.com/cznic/sqlite
