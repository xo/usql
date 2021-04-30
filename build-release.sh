#!/bin/bash

set -e

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

TAGS=(
  most
  sqlite_icu
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_stat4
  sqlite_userauth
  sqlite_vtable
  osusergo
  netgo
  static_build
)

# get latest tag version
pushd $SRC &> /dev/null
VER=$(git tag -l|grep -E '^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$'|sort -r -V|head -1||:)
popd &> /dev/null

BUILD=$SRC/build

OPTIND=1
while getopts "b:v:" opt; do
case "$opt" in
  b) BUILD=$OPTARG ;;
  v) VER=$OPTARG ;;
esac
done

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]'|sed -e 's/^\(msys\|mingw\).*/windows/')
ARCH=amd64
NAME=$(basename $SRC)
VER="${VER#v}"
EXT=tar.bz2
DIR=$BUILD/$PLATFORM/$VER
BIN=$DIR/$NAME
case $PLATFORM in
  windows)
    EXT=zip
    BIN=$BIN.exe
  ;;
  linux|darwin)
    TAGS+=(no_adodb)
  ;;
esac
OUT=$DIR/$NAME-$VER-$PLATFORM-$ARCH.$EXT

pushd $SRC &> /dev/null
echo "APP:         $NAME/${VER} ($PLATFORM/$ARCH)"
echo "BUILD TAGS:  ${TAGS[@]}"
if [ -d $DIR ]; then
  echo "REMOVING:    $DIR"
  rm -rf $DIR
fi
mkdir -p $DIR
echo "BUILDING:    $BIN"

# build parameters
TAGS="${TAGS[@]}"
EXTLDFLAGS=(
  -fno-PIC
  -static
  -licuuc
  -licui18n
  -licudata
  -ldl
)
EXTLDFLAGS="${EXTLDFLAGS[@]}"
LDFLAGS=(
  -s
  -w
  -X github.com/xo/usql/text.CommandName=$NAME
  -X github.com/xo/usql/text.CommandVersion=$VER
  -linkmode=external
  -extldflags \'$EXTLDFLAGS\'
  -extld g++
)
LDFLAGS="${LDFLAGS[@]}"

log() {
  cat - | while read -r message; do
    echo "$1$message"
  done
}

# build
(set -x;
  go build \
    -tags="$TAGS" \
    -gccgoflags="all=-DU_STATIC_IMPLEMENTATION" \
    -buildmode=pie \
    -ldflags="$LDFLAGS" \
    -o $BIN
) 2>&1 | log 'BUILDING:    '

# strip
case $PLATFORM in
  linux|windows|darwin)
    echo "STRIPPING:   $BIN"
    strip $BIN
  ;;
esac

# compress
case $PLATFORM in
  linux|windows|darwin)
    COMPRESSED=$(upx -q -q $BIN|awk '{print $1 " -> " $3 " (" $4 ")"}')
    echo "COMPRESSED:  $COMPRESSED"
  ;;
esac

# check build
BUILT_VER=$($BIN --version)
if [ "$BUILT_VER" != "$NAME ${VER#v}" ]; then
  echo -e "\n\nerror: expected $NAME --version to report '$NAME ${VER#v}', got: '$BUILT_VER'"
  exit 1
fi
echo "REPORTED:    $BUILT_VER"
case $EXT in
  tar.bz2)
    tar -C $DIR -cjf $OUT $(basename $BIN)
  ;;
  zip)
    zip $OUT -j $BIN
  ;;
esac
echo "PACKED:      $OUT ($(du -sh $OUT|awk '{print $1}'))"
popd &> /dev/null
