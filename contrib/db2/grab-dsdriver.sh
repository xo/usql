#!/bin/bash

DEST=${1:-/opt/db2}

BASE=https://raw.githubusercontent.com/owlz84/bikepoint-demo/master/drivers
FILE=ibm_data_server_driver_package_linuxx64_v11.1.tar.gz

if [ ! -w $DEST ]; then
  echo "ERROR: not able to write to $DEST"
  exit 1
fi

echo "DEST: $DEST"

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
cache $DEST $BASE/$FILE

# remove existing directory, if any
if [ -e $DEST/dsdriver ]; then
  echo "REMOVING: $DEST/dsdriver"
  rm -rf $DEST/dsdriver
fi

if [ -e $DEST/clidriver ]; then
  echo "REMOVING: $DEST/clidriver"
  rm -rf $DEST/clidriver
fi

USER=$(whoami)

# extract
pushd $DEST &> /dev/null
echo "EXTRACTING: $FILE"

# extract
tar -zxf $FILE
tar -zxf dsdriver/odbc_cli_driver/linuxamd64/ibm_data_server_driver_for_odbc_cli.tar.gz

# fix permissions
chown $USER:$USER -R .
find ./ -type d -exec chmod 0755 {} \;
find ./ -type d -exec chmod -s {} \;

popd &> /dev/null
