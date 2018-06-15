#!/bin/bash

DEST=${1:-/opt/oracle}

BASE=https://raw.githubusercontent.com/strongloop/loopback-oracle-builder/master/deps/oracle/Linux/x64
LITE=$BASE/instantclient-basiclite-linux.x64-12.1.0.2.0.zip
SDK=$BASE/instantclient-sdk-linux.x64-12.1.0.2.0.zip

if [ ! -w $DEST ]; then
  echo "ERROR: not able to write to $DEST"
  exit 1
fi

echo "DEST: $DEST"
echo "LITE: $LITE"
echo "SDK:  $SDK"

grab() {
  echo -n "RETRIEVING: $1 -> $2     "
  wget --progress=dot -O $2 $1 2>&1 |\
    grep --line-buffered "%" | \
    sed -u -e "s,\.,,g" | \
    awk '{printf("\b\b\b\b%4s", $2)}'
  echo -ne "\b\b\b\b"
  echo " DONE."
}

cache() {
  FILE=$(basename $2)
  if [ ! -f $1/$FILE ]; then
    grab $2 $1/$FILE
  fi
}

if [ ! -e "$DEST" ]; then
  echo "$DEST does not exist"
  exit 1
fi

set -e

# retrieve
cache $DEST $LITE
cache $DEST $SDK

# remove existing directory, if any
if [ -e $DEST/instantclient_12_1 ]; then
  echo "REMOVING: $DEST/instantclient_12_1"
  rm -rf $DEST/instantclient_12_1
fi

# extract
pushd $DEST &> /dev/null
unzip -qq $(basename $LITE)
unzip -qq $(basename $SDK)
popd &> /dev/null

# add missing symlinks
pushd $DEST/instantclient_12_1 &> /dev/null
ln -s libclntshcore.so.12.1 libclntshcore.so
ln -s libclntsh.so.12.1 libclntsh.so
ln -s libocci.so.12.1 libocci.so
popd &> /dev/null

# write pkg-config file
DATA=$(cat << 'ENDSTR'
prefix=${pcfiledir}

version=12.1
build=client64

libdir=${prefix}/instantclient_12_1
includedir=${prefix}/instantclient_12_1/sdk/include

Name: OCI
Description: Oracle database engine
Version: ${version}
Libs: -L${libdir} -lclntsh
Libs.private:
Cflags: -I${includedir}
ENDSTR
)
echo "$DATA" > $DEST/oci8.pc
