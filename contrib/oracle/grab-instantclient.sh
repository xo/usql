#!/bin/bash

DEST=${1:-/opt/oracle}

VERSION=19.3.0.0.0dbru
DVER=$(awk -F. '{print $1 "_" $2}' <<< "$VERSION")

BASE=https://media.githubusercontent.com/media/epoweripione/oracle-instantclient-18/master

BASIC=$BASE/instantclient-basic-linux.x64-$VERSION.zip
SDK=$BASE/instantclient-sdk-linux.x64-$VERSION.zip

if [ ! -w $DEST ]; then
  echo "ERROR: not able to write to $DEST"
  exit 1
fi

echo "DEST:  $DEST"
echo "BASIC: $BASIC"
echo "SDK:   $SDK"

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
cache $DEST $BASIC
cache $DEST $SDK

# remove existing directory, if any
if [ -e $DEST/instantclient_$DVER ]; then
  echo "REMOVING: $DEST/instantclient_$DVER"
  rm -rf $DEST/instantclient_$DVER
fi

# extract
pushd $DEST &> /dev/null
unzip -qq $(basename $BASIC)
unzip -qq $(basename $SDK)
popd &> /dev/null

# write pkg-config file
DATA=$(cat <<ENDSTR
prefix=\${pcfiledir}

version=$VERSION
build=client64

libdir=\${prefix}/instantclient_${DVER}
includedir=\${prefix}/instantclient_${DVER}/sdk/include

Name: OCI
Description: Oracle database engine
Version: ${VERSION}
Libs: -L\${libdir} -lclntsh
Libs.private:
Cflags: -I\${includedir}
ENDSTR
)
echo "$DATA" > $DEST/oci8.pc
