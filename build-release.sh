#!/bin/bash

set -e

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

VER=
BUILD=$SRC/build
STATIC=0
FORCE=0
CHECK=1
VERBOSE=false
PLATFORM=$(go env GOOS)
ARCH=$(go env GOARCH)
GOARCH=$ARCH

OPTIND=1
while getopts "a:b:v:sfrnx" opt; do
case "$opt" in
  a) ARCH=$OPTARG ;;
  b) BUILD=$OPTARG ;;
  v) VER=$OPTARG ;;
  s) STATIC=1 ;;
  f) FORCE=1 ;;
  n) CHECK=0 ;;
  r)
    # get latest tag version
    pushd $SRC &> /dev/null
    VER=$(git tag -l|grep -E '^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$'|sort -r -V|head -1||:)
    popd &> /dev/null
  ;;
  x) VERBOSE=true ;;
esac
done

# neither -v or -r specified, set FORCE and VER
if [ "$VER" = "" ]; then
  VER=0.0.0-dev
  FORCE=1
fi

NAME=$(basename $SRC)
VER="${VER#v}"
EXT=tar.bz2
DIR=$BUILD/$PLATFORM/$ARCH/$VER
BIN=$DIR/$NAME

TAGS=(
  most
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_math_functions
  sqlite_stat4
  sqlite_userauth
  sqlite_vtable
)
case $PLATFORM in
  darwin|linux)
    TAGS+=(no_adodb)
  ;;
  windows)
    EXT=zip
    BIN=$BIN.exe
  ;;
esac
OUT=$DIR/$NAME-$VER-$PLATFORM-$ARCH.$EXT

CARCH=
QEMUARCH=
GNUTYPE=
CC=
CXX=
EXTLD=g++

if [[ "$PLATFORM" == "linux" && "$ARCH" != "$GOARCH" ]]; then
  case $ARCH in
    arm)   CARCH=armhf   QEMUARCH=arm     GNUTYPE=gnueabihf ;;
    arm64) CARCH=aarch64 QEMUARCH=aarch64 GNUTYPE=gnu ;;
    *)
      echo "error: unknown arch $ARCH"
      exit 1
    ;;
  esac
  LDARCH=$CARCH
  if [[ "$ARCH" == "arm" ]]; then
    TAGS+=(no_netezza)
    if [ -d /usr/arm-linux-$GNUTYPE ]; then
      LDARCH=arm
    elif [ -d /usr/arm-none-linux-$GNUTYPE ]; then
      LDARCH=arm-none
    fi
  fi
  CC=$LDARCH-linux-$GNUTYPE-gcc
  CXX=$LDARCH-linux-$GNUTYPE-c++
  EXTLD=$LDARCH-linux-$GNUTYPE-g++
fi

LDFLAGS=(
  -s
  -w
  -X github.com/xo/usql/text.CommandName=$NAME
  -X github.com/xo/usql/text.CommandVersion=$VER
)

if [ "$STATIC" = "1" ]; then
  OUT=$DIR/${NAME}_static-$VER-$PLATFORM-$ARCH.$EXT
  BIN=$DIR/${NAME}_static
  case $PLATFORM in
    linux)
      TAGS+=(
        netgo
        osusergo
      )
      EXTLDFLAGS=(
        -static
        -lm
        -ldl
      )
      EXTLDFLAGS="${EXTLDFLAGS[@]}"
      LDFLAGS+=(
        -linkmode=external
        -extldflags \'$EXTLDFLAGS\'
        -extld $EXTLD
      )
    ;;
    *)
      echo "ERROR: fully static builds not currently supported for $PLATFORM/$ARCH"
      exit 1
    ;;
  esac
fi

# check not overwriting existing build artifacts
if [[ -e $OUT && "$FORCE" != "1" ]]; then
  echo "ERROR: $OUT exists and FORCE != 1 (try $0 -f)"
  exit 1
fi

TAGS="${TAGS[@]}"
LDFLAGS="${LDFLAGS[@]}"

log() {
  cat - | while read -r message; do
    echo "$1$message"
  done
}

echo "APP:         $NAME/${VER} ($PLATFORM/$ARCH)"
if [ "$STATIC" = "1" ]; then
  echo "STATIC:      yes"
fi
echo "BUILD TAGS:  $TAGS"
echo "LDFLAGS:     $LDFLAGS"

pushd $SRC &> /dev/null

if [ -f $OUT ]; then
  echo "REMOVING:    $OUT"
  rm -rf $OUT
fi
mkdir -p $DIR
echo "BUILDING:    $BIN"

# build
echo "BUILD:"

(set -x;
  CC=$CC \
  CXX=$CXX \
  CGO_ENABLED=1 \
  GOARCH=$ARCH \
  go build \
    -v=$VERBOSE \
    -x=$VERBOSE \
    -ldflags="$LDFLAGS" \
    -tags="$TAGS" \
    -trimpath \
    -o $BIN
) 2>&1 | log '    '

built_ver() {
  if [[ "$ARCH" != "$GOARCH" ]]; then
    EXTRA=
    if [ -d /usr/$LDARCH-linux-$GNUTYPE/libc ]; then
      EXTRA="-L /usr/$LDARCH-linux-$GNUTYPE/libc"
    fi
    qemu-$QEMUARCH \
      -L /usr/$LDARCH-linux-$GNUTYPE \
      $EXTRA \
      $BIN --version
  else
    $BIN --version
  fi
}

# check build
if [[ "$CHECK" == "1" ]]; then
  BUILT_VER=$(built_ver)
  if [ "$BUILT_VER" != "$NAME ${VER#v}" ]; then
    echo -e "\n\nERROR: expected $NAME --version to report '$NAME ${VER#v}', got: '$BUILT_VER'"
    exit 1
  fi
  echo "REPORTED:    $BUILT_VER"
fi

# pack
cp $SRC/LICENSE $DIR
case $EXT in
  tar.bz2)
    tar -C $DIR -cjf $OUT $(basename $BIN) LICENSE
  ;;
  zip)
    zip $OUT -j $BIN LICENSE
  ;;
esac

# report
echo "PACKED:      $OUT ($(du -sh $OUT|awk '{print $1}'))"
case $EXT in
  tar.bz2) tar -jvtf $OUT ;;
  zip)     unzip -l  $OUT ;;
esac

popd &> /dev/null
