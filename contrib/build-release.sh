#!/bin/bash

VER=$1
BUILD=$2
EXT=tar.bz2

if [ -z "$VER" ]; then
  echo "usage: $0 <VER>"
  exit 1
fi

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]')

TAG=v$VER
SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)
NAME=$(basename $SRC)

if [ -z "$BUILD" ]; then
  BUILD=$SRC/build
fi

DIR=$BUILD/$PLATFORM/$VER
BIN=$DIR/$NAME
OUT=$DIR/usql-$VER-$PLATFORM-amd64.$EXT

case $PLATFORM in
  mingw64)
    PLATFORM=windows
    EXT=zip
    BIN=$BIN.exe
  ;;
  msys)
    PLATFORM=windows
    EXT=zip
    BIN=$BIN.exe
  ;;
esac

echo "PLATFORM: $PLATFORM"
echo "VER: $VER"
echo "DIR: $DIR"
echo "BIN: $BIN"
echo "OUT: $OUT"

if [ -d $DIR ]; then
  echo "removing existing $DIR"
  rm -rf $DIR
fi

mkdir -p $DIR

set -e

pushd $SRC &> /dev/null

go build -ldflags="-X main.name=$NAME -X main.version=$VER" -o $BIN

echo -n "checking usql --version: "
BUILT_VER=$($BIN --version)
if [ "$BUILT_VER" != "usql $VER" ]; then
  echo -e "\n\nexpected --version to be 'usql $VER', got: '$BUILT_VER'"
  exit 1
fi
echo "$BUILT_VER"

if [ "$PLATFORM" == "linux" ]; then
  echo "stripping $BIN"
  strip $BIN

  echo "packing $BIN"
  upx -q -q $BIN
fi

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
