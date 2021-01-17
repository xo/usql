#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../)
ALL=$(find $SRC/drivers/ -mindepth 1 -maxdepth 1 -type d|sort|grep -v genji)

PKGS=
for i in $ALL; do
  NAME=$(basename $i)
  PKG=$(sed -n '/DRIVER: /p' $i/$NAME.go |sed -e 's/^\(\s\|"\|_\)\+//'|sed -e 's/[a-z]\+\s\+"//' |sed -e 's/".*//')
  PKGS="$PKGS $PKG"
done

set -e -x

go get $@ $PKGS
