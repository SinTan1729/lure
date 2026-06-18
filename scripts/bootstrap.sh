#!/bin/bash

# LURE - Linux User REpository
# Copyright (C) 206 Sayantan Santra
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

info() {
    echo $'\x1b[32m[INFO]\x1b[0m' $@
}

warn() {
    echo $'\x1b[31m[WARN]\x1b[0m' $@
}

error() {
    echo $'\x1b[31;1m[ERR]\x1b[0m' $@
    exit 1
}

installPkg() {
    rootCmd=""
    if command -v doas &>/dev/null; then
        rootCmd="doas"
    elif command -v sudo &>/dev/null; then
        rootCmd="sudo"
    else
        warn "No privilege elevation command (e.g. sudo, doas) detected"
    fi

    case $1 in
    pacman) $rootCmd pacman --noconfirm -U ${@:2} ;;
    apk) $rootCmd apk add --allow-untrusted ${@:2} ;;
    zypper) $rootCmd zypper --no-gpg-checks install ${@:2} ;;
    *) $rootCmd $1 install -y ${@:2} ;;
    esac
}

if ! command -v curl &>/dev/null; then
    error "This script requires the curl command. Please install it and run again."
fi

latestVersion=$(curl -sI 'https://github.com/SinTan1729/lure/releases/latest' | grep -io 'location: .*' | rev | cut -d '/' -f1 | rev | tr -d '[:space:]')
info "Found latest LURE version:" $latestVersion

arch=$(uname -m)
case $arch in
armv*) arch="arm" ;;
i686) arch="i386" ;;
esac

tmpdir=$(mktemp -d -t lure-bootstrap.XXXXXXX)
cd $tmpdir
# Use ${arch} instead of $(uname -m)
filename="lure-${latestVersion}-linux-${arch}"
url="https://github.com/Sintan1729/lure/releases/download/${latestVersion}/${filename}.tar.gz"
echo $url

info "Downloading LURE package"
curl -L $url -o lure.tar.gz
tar -xzf lure.tar.gz -C .

info "Installing LURE package"
mv $filename lure
./lure install linux-user-repository-bin

info "Cleaning up"
rm -rf $tmpdir

info "Done!"
