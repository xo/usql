#!/bin/bash

# most.sh builds a development version of usql with "most" drivers.
#
# Options:
#  -i        go install instead of go build
#  -m        disable sqlite3 driver (no_sqlite3) and diasble CGO.
#            causes moderncsqlite to register aliases for sqlite3.
#  -v        add -v to go build/install
#  -x        add -x to go build/install
#  -t <TAG>  add build tag
#
# Usage: ./most.sh -i -t godror -v

NAME=usql
VER="$(date +%y.%m.%d)-dev"

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]'|sed -e 's/^\(msys\|mingw\).*/windows/')

CGO_ENABLED=1
BUILDVERB=build
TAGS=()
SQLITE_TAGS=(
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_stat4
  sqlite_userauth
  sqlite_vtable
)
EXTRA=()

case $PLATFORM in
  darwin|linux)
    TAGS+=(no_adodb)
    SQLITE_TAGS+=(sqlite_icu)
  ;;
esac

OPTIND=1
while getopts "imvxt:" opt; do
case "$opt" in
  i) BUILDVERB=install ;;
  m) CGO_ENABLED=0 SQLITE_TAGS=(no_sqlite3) ;;
  v) EXTRA+=(-v) ;;
  x) EXTRA+=(-x) ;;
  t) TAGS+=($OPTARG) ;;
esac
done

TAGS="most ${SQLITE_TAGS[@]} ${TAGS[@]}"
LDFLAGS=(
  -s
  -w
  -X github.com/xo/usql/text.CommandName=$NAME
  -X github.com/xo/usql/text.CommandVersion=$VER
)
LDFLAGS="${LDFLAGS[@]}"

(set -x;
  CGO_ENABLED=$CGO_ENABLED \
    go $BUILDVERB \
    -tags="$TAGS" \
    -ldflags="$LDFLAGS" \
    ${EXTRA[@]}
)
