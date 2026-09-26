#!/bin/bash

set -e

# Note: To be able to run this on a Debian/Ubuntu system, ensure you have these
# dependencies installed:
#   apt install -y golang ca-certificates file bzip2

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

NAME=$(basename $SRC)
VER=
STATIC=0
FORCE=0
CHECK=1
INSTALL=0
BUILDONLY=0
VERBOSE=false
PRINTTAGS=0
CGO_ENABLED=1
LDNAME=github.com/xo/usql/text.CommandName
LDVERSION=github.com/xo/usql/text.CommandVersion
PLATFORM=$(go env GOOS)
ARCH=$(go env GOARCH)
GOARCH=$ARCH

TAGS=(
  most
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_math_functions
  sqlite_stat4
  sqlite_vtable
)

latest_tag() {
  # get latest tag version
  pushd $SRC &> /dev/null
  git tag -l|grep -E '^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$'|sort -r -V|head -1||:
  popd &> /dev/null
}

usage() {
  cat <<'END'
usage: build.sh [options]

Builds usql. With no options it builds version 0.0.0-dev for the host system,
packs it, and checks that the binary reports the version it was built with.

options:
  -a, --arch ARCH     build for ARCH instead of the host architecture
  -v, --version VER   build version VER, instead of 0.0.0-dev
  -r, --release       build the most recent version tag
  -t, --tags 'TAGS'   build with TAGS instead of the default set
  -s, --static        build a fully static binary (linux only)
  -b, --build-only    build only: do not write, check or pack a binary
  -i, --install       run go install instead of go build
  -n, --no-check      do not check that the binary reports its version
  -f, --force         overwrite an existing archive
  -x, --verbose       pass -v -x to go build
  -T, --print-tags    print the build tags that would be used, then stop
  -h, --help          show this text

examples:
  build.sh -b                     build for the host, keep nothing
  build.sh -r                     build and pack the most recent tag
  build.sh -a arm64 -v 1.2.3      cross build version 1.2.3 for arm64
  build.sh -t 'all' -b            build with every driver
  go test -tags "$(build.sh -T)"  test with the same tags as the build
END
}

# Translate long options to their short forms, because getopts reads short
# options alone. Anything else is passed through untouched.
ARGS=()
while [ $# -ne 0 ]; do
  case "$1" in
    --arch)        ARGS+=(-a "$2"); shift 2 ;;
    --version)     ARGS+=(-v "$2"); shift 2 ;;
    --tags)        ARGS+=(-t "$2"); shift 2 ;;
    --release)     ARGS+=(-r); shift ;;
    --static)      ARGS+=(-s); shift ;;
    --build-only)  ARGS+=(-b); shift ;;
    --install)     ARGS+=(-i); shift ;;
    --no-check)    ARGS+=(-n); shift ;;
    --force)       ARGS+=(-f); shift ;;
    --verbose)     ARGS+=(-x); shift ;;
    --print-tags)  ARGS+=(-T); shift ;;
    --help)        ARGS+=(-h); shift ;;
    --*)
      echo "error: unknown option $1" >&2
      usage >&2
      exit 1
    ;;
    *) ARGS+=("$1"); shift ;;
  esac
done
set -- ${ARGS[@]+"${ARGS[@]}"}

OPTIND=1
while getopts "a:v:sfnibxt:rTh" opt; do
case "$opt" in
  a) ARCH=$OPTARG ;;
  v) VER=$OPTARG ;;
  s) STATIC=1 ;;
  f) FORCE=1 ;;
  n) CHECK=0 ;;
  i) INSTALL=1 ;;
  b) BUILDONLY=1 ;;
  x) VERBOSE=true ;;
  t) TAGS=($OPTARG) ;;
  r) VER=$(latest_tag) ;;
  T) PRINTTAGS=1 ;;
  h) usage; exit 0 ;;
  *) usage >&2; exit 1 ;;
esac
done

# neither -v or -r specified, or -v=master, set FORCE and VER
if [[ "$VER" = "" || "$VER" == "master" ]]; then
  VER=0.0.0-dev
  FORCE=1
fi

VER="${VER#v}"

BUILD=$SRC/build
DIR=$BUILD/$PLATFORM/$ARCH/$VER

TAR=tar
EXT=tar.bz2
BIN=$DIR/$NAME

case $PLATFORM in
  linux)
    TAGS+=(no_adodb)
  ;;
  windows)
    EXT=zip
    BIN=$BIN.exe
  ;;
  darwin)
    TAGS+=(no_adodb)
    TAR=gtar
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
    TAGS+=(no_netezza no_chai)
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
  -X $LDNAME=$NAME
  -X $LDVERSION=$VER
)

if [ "$STATIC" = "1" ]; then
  OUT=$DIR/${NAME}_static-$VER-$PLATFORM-$ARCH.$EXT
  BIN=$DIR/${NAME}_static
  case $PLATFORM in
    linux)
      TAGS+=(
        netgo
        osusergo
        minicore_disabled
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

# report resolved build tags (used by update-deps.sh)
if [ "$PRINTTAGS" = "1" ]; then
  echo "${TAGS[@]}"
  exit
fi

# check not overwriting existing build artifacts
if [[ -e $OUT && "$FORCE" != "1" && "$INSTALL" == "0" ]]; then
  echo "ERROR: $OUT exists and FORCE != 1 (try $0 -f)"
  exit 1
fi

TAGS="${TAGS[@]}"
LDFLAGS="${LDFLAGS[@]}"

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

VERB=build
OUTPUT="-o $BIN"
if [ "$INSTALL" = "1" ]; then
  VERB=install OUTPUT=""
elif [ "$BUILDONLY" = "1" ]; then
  OUTPUT=""
fi
(set -x;
  CC=$CC \
  CXX=$CXX \
  CGO_ENABLED=$CGO_ENABLED \
  GOARCH=$ARCH \
  go $VERB \
    -v=$VERBOSE \
    -x=$VERBOSE \
    -ldflags="$LDFLAGS" \
    -tags="$TAGS" \
    -trimpath \
    $OUTPUT
)

# Install if requested
if [ "$INSTALL" = "1" ]; then
  echo "INSTALLING BINARY"
  (set -x;
    go install \
      -v=$VERBOSE \
      -x=$VERBOSE \
      -ldflags="$LDFLAGS" \
      -tags="$TAGS" \
      -trimpath
  )


  if [[ "$PLATFORM" == "windows" ]]; then
    echo "SKIP MAN PAGE ON WINDOWS"
    exit 0
  fi

  # Install man page on non-Windows systems
  MANDIR=${MANDIR:-/usr/local/share/man/man1}
  echo "INSTALLING MAN PAGE to $MANDIR"
  (set -x;
    mkdir -p "$MANDIR"
    install -m 0644 "$SRC/man/usql.1" "$MANDIR/"
  )
  # Update man database
  if command -v mandb &>/dev/null; then
    echo "UPDATING MAN DATABASE"
    (set -x; mandb)
  fi
  exit 0
fi

# Exit if only building
if [ "$BUILDONLY" = "1" ]; then
    exit 0
fi

(set -x;
  file $BIN
)
if [[ "$PLATFORM" != "windows" ]]; then
  (set -x;
    chmod +x $BIN
  )
fi

# purge disk cache
if [[ "$PLATFORM" == "darwin" && "$CI" == "true" ]]; then
  (set -x;
    sudo /usr/sbin/purge
  )
fi

built_ver() {
  if [[ "$PLATFORM" == "linux" && "$ARCH" != "$GOARCH" ]]; then
    EXTRA=
    if [ -d /usr/$LDARCH-linux-$GNUTYPE/libc ]; then
      EXTRA="-L /usr/$LDARCH-linux-$GNUTYPE/libc"
    fi
    qemu-$QEMUARCH \
      -L /usr/$LDARCH-linux-$GNUTYPE \
      $EXTRA \
      $BIN --version
  elif [[ "$PLATFORM" == "darwin" && "$ARCH" != "$GOARCH" ]]; then
    echo "$NAME ${VER#v}"
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
cp $SRC/LICENSE $DIR/
cp $SRC/man/usql.1 $DIR/
case $EXT in
  tar.bz2) $TAR -C $DIR -cjf $OUT $(basename $BIN) LICENSE usql.1 ;;
  zip) zip $OUT -j $BIN $DIR/LICENSE $DIR/usql.1 ;;
esac

# report
echo "PACKED:      $OUT ($(du -sh $OUT|awk '{print $1}'))"

case $EXT in
  tar.bz2) (set -x; $TAR  -jvtf $OUT) ;;
  zip)     (set -x; unzip -l    $OUT) ;;
esac

(set -x;
  sha256sum $DIR/$(basename $BIN) $DIR/LICENSE $DIR/usql.1
)

popd &> /dev/null
