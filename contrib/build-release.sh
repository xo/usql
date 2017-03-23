#!/bin/bash

VER=$1
BUILD=$2

if [ -z "$VER" ]; then
  echo "usage: $0 <VER>"
  exit 1
fi

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]')
TAG=v$VER
SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)
NAME=$(basename $SRC)
EXT=tar.bz2

if [ -z "$BUILD" ]; then
  BUILD=$SRC/build
fi

DIR=$BUILD/$PLATFORM/$VER
BIN=$DIR/$NAME

DATABASES="avatica clickhouse couchbase firebird mymysql pgx ql saphana voltdb yql"
EXTRA="icu fts5 vtable json1"

case $PLATFORM in
  mingw64|msys)
    PLATFORM=windows
    EXT=zip
    BIN=$BIN.exe
    DATABASES="$DATABASES adodb"

    ICU=icu-i18n-mingw64
    CGO_CFLAGS="$(go env CGO_CFLAGS) $(pkg-config --cflags $ICU)"
    CGO_LDFLAGS="$(go env CGO_LDFLAGS) $(pkg-config --libs-only-L $ICU)"
  ;;

  darwin)
    ICU=icu-i18n
    CGO_CFLAGS="$(go env CGO_CFLAGS) $(pkg-config --cflags $ICU)"
    CGO_LDFLAGS="$(go env CGO_LDFLAGS) $(pkg-config --libs-only-L $ICU)"
  ;;
esac

OUT=$DIR/usql-$VER-$PLATFORM-amd64.$EXT

echo "PLATFORM: $PLATFORM"
echo "VER: $VER"
echo "DIR: $DIR"
echo "BIN: $BIN"
echo "OUT: $OUT"
echo "DATABASES: $DATABASES"
echo "EXTRA: $EXTRA"
echo "CGO_CFLAGS: $CGO_CFLAGS"
echo "CGO_LDFLAGS: $CGO_LDFLAGS"

set -e

if [ -d $DIR ]; then
  echo "removing $DIR"
  rm -rf $DIR
fi

mkdir -p $DIR

pushd $SRC &> /dev/null

CGO_CFLAGS=$CGO_CFLAGS \
CGO_LDFLAGS=$CGO_LDFLAGS \
  go build \
    -tags "$DATABASES $EXTRA" \
    -ldflags="-X github.com/knq/usql/text.CommandName=$NAME -X github.com/knq/usql/text.CommandVersion=$VER" \
    -o $BIN

echo -n "checking usql --version: "
BUILT_VER=$($BIN --version)
if [ "$BUILT_VER" != "usql $VER" ]; then
  echo -e "\n\nerror: expected usql --version to report 'usql $VER', got: '$BUILT_VER'"
  exit 1
fi
echo "$BUILT_VER"

case $PLATFORM in
  linux|windows|darwin)
    echo "stripping $BIN"
    strip $BIN
  ;;
esac

case $PLATFORM in
  linux|windows)
    echo "packing $BIN"
    upx -q -q $BIN
  ;;
esac

echo "compressing $OUT"
case $EXT in
  tar.bz2)
    tar -C $DIR -cjf $OUT $(basename $BIN)
  ;;
  zip)
    zip $OUT -j $BIN
  ;;
esac

du -sh $OUT

popd &> /dev/null
