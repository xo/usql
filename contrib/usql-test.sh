#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

USQL=$(which usql)
if [ -f $SRC/../usql ]; then
  USQL=$(realpath $SRC/../usql)
fi

export USQL_SHOW_HOST_INFORMATION=false
for TARGET in $SRC/*/usql-config; do
  NAME=$(basename $(dirname $TARGET))
  if [[ ! -z "$(podman ps -q --filter "name=$NAME")" || "$NAME" == "duckdb" || "$NAME" == "sqlite3" ]]; then
    unset DB VSQL
    source $TARGET
    if [[ -z "$DB" || -z "$VSQL" ]]; then
      echo -e "ERROR: DB or VSQL not defined in $TARGET!\n"
      continue
    fi
    (set -x;
      $USQL "$DB" -X -J -c "$VSQL"
    )
    echo
  fi
done
