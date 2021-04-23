#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

USQL=$(which usql)
if [ -f $SRC/../usql ]; then
  USQL=$(realpath $SRC/../usql)
fi

for TARGET in $SRC/*/usql-config; do
  NAME=$(basename $(dirname $TARGET))
  if [ ! -z "$(docker ps -q --filter "name=$NAME")" ]; then
    unset DB VSQL
    source $TARGET
    if [ -z "$DB" ]; then
      echo "error: DB not defined in $TARGET/usql-config!"
      exit 1
    fi
    if [ -z "$VSQL" ]; then
      echo "error: VSQL not defined in $TARGET/usql-config!"
      exit 1
    fi
    (set -x;
      USQL_SHOW_HOST_INFORMATION=false \
        $USQL "$DB" -X -J -c "$VSQL"
    )
  fi
done
