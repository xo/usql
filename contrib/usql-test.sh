#!/bin/bash

DIR=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

BASE=$SRC/$DIR

if [ -z "$DIR" ]; then
  echo "usage: $0 <NAME>"
  exit 1
fi

if [ ! -e $BASE/usql-config ]; then
  echo "error: $BASE/usql-config doesn't exist"
  exit 1
fi

source $BASE/usql-config

if [ -z "$DB" ]; then
  echo "error: DB not defined in $BASE/usql-config!"
  exit 1
fi

USQL=$(realpath $SRC/../usql)
USQL_SHOW_HOST_INFORMATION=false \
  $USQL "$DB" -X -J -c "$VSQL"
