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

case $PLATFORM in
  mingw64|msys)
    PLATFORM=windows
    EXT=zip
    BIN=$BIN.exe
  ;;
esac

OUT=$DIR/usql-$VER-$PLATFORM-amd64.$EXT

echo "PLATFORM: $PLATFORM"
echo "VER: $VER"
echo "DIR: $DIR"
echo "BIN: $BIN"
echo "OUT: $OUT"

set -e

if [ -d $DIR ]; then
  echo "removing $DIR"
  rm -rf $DIR
fi

mkdir -p $DIR

pushd $SRC &> /dev/null

TAGS=
case $PLATFORM in
  windows)
    TAGS="-tags adodb"
  ;;
esac

go build -ldflags="-X github.com/knq/usql/text.CommandName=$NAME -X github.com/knq/usql/text.CommandVersion=$VER" $TAGS -o $BIN

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
