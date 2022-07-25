#!/bin/bash

DEST=${1:-/opt/oracle}

# available versions:
# 21.7.0.0.0
# 21.6.0.0.0
# 21.1.0.0.0
# 19.9.0.0.0
# 18.5.0.0.0
# 12.2.0.1.0

VERSION=

OPTIND=1
while getopts "v:" opt; do
case "$opt" in
  v) VERSION=$OPTARG ;;
esac
done

if [ -z "$VERSION" ]; then
  VERSION=$(
    wget --quiet -O- https://www.oracle.com/database/technologies/instant-client/linux-x86-64-downloads.html| \
      sed -n -e 's/.*\/instantclient-basic-linux\.x64-\([^d]\+\)dbru\.zip.*/\1/p' | \
      head -1
    )
fi

if [[ ! "$VERSION" =~ ^[0-9\.]+$ ]]; then
  echo "error: invalid VERSION"
  exit 1
fi

BASE=https://download.oracle.com/otn_software/linux/instantclient/$(sed -e 's/[^0-9]//g' <<< "$VERSION")

# build list of archives to retrieve
declare -a ARCHIVES
for i in basic sdk sqlplus; do
  ARCHIVES+=("$BASE/instantclient-$i-linux.x64-${VERSION}dbru.zip")
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
rm -f /etc/ld.so.conf.d/oracle-instantclient.conf
echo "$DEST/instantclient_$DVER" | tee -a /etc/ld.so.conf.d/oracle-instantclient.conf
ldconfig -v

# write sqlnet.ora
DATA=$(cat <<ENDSTR
DIAG_ADR_ENABLED = OFF
TRACE_LEVEL_CLIENT = OFF
TRACE_DIRECTORY_CLIENT = /dev/null
LOG_DIRECTORY_CLIENT = /dev/null
LOG_FILE_CLIENT = /dev/null
LOG_LEVEL_CLIENT = OFF
ENDSTR
)
echo "$DATA" > $DEST/instantclient_${DVER}/network/admin/sqlnet.ora
