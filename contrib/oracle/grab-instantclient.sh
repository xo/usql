#!/bin/bash

DEST=${1:-/opt/oracle}

# https://download.oracle.com/otn_software/linux/instantclient/193000/instantclient-basic-linux.x64-19.3.0.0.0dbru.zip
VERSION=19.3.0.0.0dbru
BASE=https://download.oracle.com/otn_software/linux/instantclient/193000

# build list of archives to retrieve
declare -a ARCHIVES
for i in basic sdk sqlplus; do
  ARCHIVES+=("$BASE/instantclient-$i-linux.x64-$VERSION.zip")
done

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

set -e

echo "DEST:       $DEST"
if [ ! -w $DEST ]; then
  echo "$DEST is not writable!"
  exit 1
fi
if [ ! -e "$DEST" ]; then
  echo "$DEST does not exist!"
  exit 1
fi

# retrieve archives
for i in ${ARCHIVES[@]}; do
  cache $DEST $i
done

# remove existing directory, if any
DVER=$(awk -F. '{print $1 "_" $2}' <<< "$VERSION")
if [ -e $DEST/instantclient_$DVER ]; then
  echo "REMOVING:   $DEST/instantclient_$DVER"
  rm -rf $DEST/instantclient_$DVER
fi

# extract
pushd $DEST &> /dev/null
for i in ${ARCHIVES[@]}; do
  unzip -qq $(basename $i)
done
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
