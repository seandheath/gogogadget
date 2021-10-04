#!/bin/bash

declare -A za
# Convert Go architectures to Zig
za['aix/ppc64']=''
za['android/386']=''
za['android/amd64']=''
za['android/arm']=''
za['android/arm64']=''
za['darwin/amd64']=''
za['darwin/arm64']=''
za['dragonfly/amd64']=''
za['freebsd/386']=''
za['freebsd/amd64']=''
za['freebsd/arm']=''
za['freebsd/arm64']=''
za['illumos/amd64']=''
za['ios/amd64']='x86_64-macos-gnu'
za['ios/arm64']=''
za['js/wasm']=''
za['linux/386']='i386-linux-gnu'
za['linux/amd64']='x86_64-linux-gnu'
za['linux/arm']='arm-linux-gnueabi'
za['linux/arm64']='aarch64-linux-gnu'
za['linux/mips']='mips-linux-gnu'
za['linux/mips64']='mips64-linux-gnu'
za['linux/mips64le']='mips64el-linux-gnueabi64'
za['linux/mipsle']='mipsel-linux-gnu'
za['linux/ppc64']='powerpc64-linux-gnu'
za['linux/ppc64le']='powerpc64le-linux-gnu'
za['linux/riscv64']='riscv64-linux-gnu'
za['linux/s390x']='s390x-linux-gnu'
za['netbsd/386']=''
za['netbsd/amd64']=''
za['netbsd/arm']=''
za['netbsd/arm64']=''
za['openbsd/386']=''
za['openbsd/amd64']=''
za['openbsd/arm']=''
za['openbsd/arm64']=''
za['openbsd/mips64']=''
za['plan9/386']=''
za['plan9/amd64']=''
za['plan9/arm']=''
za['solaris/amd64']=''
za['windows/386']='i386-windows-gnu'
za['windows/amd64']='x86_64-windows-gnu'
za['windows/arm']='arm-windows-gnu'

help() {
    echo "Select what architecture you want to build for (or leave empty)"
    echo "Available OS/Arch:"
    go tool dist list
    echo "Usage: ./compile.sh [OS ARCH]"
}

trap 'help' ERR

if [[ ! -e "build" ]]; then
    mkdir build
fi

ARGC=$#
GOOS=""
GOARCH=""
FNAME="gogogadget"
if [ $ARGC -gt "0" ]; then
    if [ $ARGC -eq 2 ]; then
        GOOS=$1
        GOARCH=$2
        FNAME="$FNAME$GOOS$GOARCH"
        echo "Building GoGoGadget for $GOOS/$GOARCH"
    else 
        help
    fi
fi

env CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -a -o build/$FNAME
