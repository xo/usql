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

PLAT=$PLATFORM
case $PLAT in
  mingw64)
    PLAT=windows
  ;;
esac

if [ -z "$BUILD" ]; then
  BUILD=$SRC/build
  if [ "$PLATFORM" == "mingw64" ]; then
    BUILD=$HOME/$NAME
  fi
fi

EXT=tar
if [ "$PLATFORM" == "mingw64" ]; then
  EXT=zip
fi

DIR=$BUILD/$PLATFORM/$VER
BIN=$DIR/$NAME
OUT=$DIR/usql-$VER-$PLAT-amd64.$EXT

rm -rf $DIR
mkdir -p $DIR

if [ "$PLATFORM" == "mingw64" ]; then
    BIN=$BIN.exe
fi

echo "PLATFORM: $PLATFORM"
echo "VER: $VER"
echo "DIR: $DIR"
echo "BIN: $BIN"
echo "OUT: $OUT"

set -e

pushd $SRC &> /dev/null

if [ "$PLATFORM" != "mingw64" ]; then
  git checkout $TAG
fi

go build -ldflags="-X main.name=$NAME -X main.version=$VER" -o $BIN

if [ "$PLATFORM" != "mingw64" ]; then
  echo "stripping $BIN"
  strip $BIN
fi

echo "packing $BIN"
upx -q -q $BIN

echo "compressing $OUT"
case $EXT in
  zip)
    zip $OUT -j $BIN
  ;;
  tar)
    tar -C $DIR -cjvf $OUT.bz2 $(basename $BIN)
  ;;
esac

if [ "$PLATFORM" == "mingw64" ]; then
  cp $OUT 'f:\'
fi

popd &> /dev/null
