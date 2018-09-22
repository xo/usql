#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)

ALL=$(find $SRC/drivers/ -mindepth 1 -maxdepth 1 -type d|sort)

SED=sed
if [ "$(uname)" == "Darwin" ]; then
  SED=gsed
fi

PKGS=
for i in $ALL; do
  NAME=$(basename $i)
  PKG=$($SED -n '/DRIVER: /{n;p;}' $i/$NAME.go|$SED -e 's/^\(\s\|"\|_\)\+//'|$SED -e 's/[a-z]\+\s\+"//' |$SED -e 's/"\s*//')
  PKGS="$PKGS $PKG"
done

set -e -x

go get $@ $PKGS
