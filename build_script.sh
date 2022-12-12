#!/bin/bash

if [ $# -ne 3 ]; 
    then echo "$0 <program.go> <program name> <output dir>"
    exit 1
fi

declare -a architecture64=("aix/ppc64" "android/amd64" "android/arm64" "darwin/amd64" "darwin/arm64" "dragonfly/amd64" "freebsd/amd64" "freebsd/arm64" "illumos/amd64" "ios/amd64" "ios/arm64" "js/wasm" "linux/amd64" "linux/arm64" "linux/mips64" "linux/mips64le" "linux/ppc64" "linux/ppc64le" "linux/riscv64" "linux/s390x" "netbsd/amd64" "netbsd/arm64" "openbsd/amd64" "openbsd/arm64" "openbsd/mips64" "plan9/amd64" "solaris/amd64" "windows/amd64" "windows/arm64")

#declare -a architecture64=("darwin/amd64")


for t in ${architecture64[@]}; do

# env CGO_ENABLED=1 GOOS=darwin GOARCH=arm64

  IFS='/'

  #Read the split words into an array based on space delimiter
  read -a strarr <<< "$t"

  #compilingstring="env CGO_ENABLED=1 GOOS=${strarr[0]} GOARCH=${strarr[1]} go build -o $3/$2_${strarr[0]}-${strarr[1]} $1"
  compilingstring="env GOOS=${strarr[0]} GOARCH=${strarr[1]} go build -o $3/$2_${strarr[0]}-${strarr[1]} $1"
  echo "Compiling $compilingstring"

  bash -c "$compilingstring 2>/dev/null"

  if [ $? -ne 0 ]; then 
    echo "[-] Cannot compile $t"
  fi
  
done