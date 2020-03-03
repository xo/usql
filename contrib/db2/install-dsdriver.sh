#!/bin/bash

DEST=${1:-/opt/db2}
FILE=$2

if [ ! -w $DEST ]; then
  echo "ERROR: not able to write to $DEST"
  exit 1
fi

echo "DEST: $DEST"

if [ ! -e "$DEST" ]; then
  echo "$DEST does not exist"
  exit 1
fi

if [ -z "$FILE" ]; then
  FILE=$(ls $HOME/Downloads/ibm_data_server_driver_package_linuxx64_*.tar.gz||:)
fi

if [ -z "$FILE" ]; then
  echo "cannot find driver package to extract"
  exit 1
fi

set -e

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
