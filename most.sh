#!/bin/bash

NAME=usql
VER="$(date +%y.%m.%d)-dev"

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]'|sed -e 's/^\(msys\|mingw\).*/windows/')

TAGS=(
  most
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_stat4
  sqlite_userauth
  sqlite_vtable
)

case $PLATFORM in
  darwin|linux)
    TAGS+=(sqlite_icu no_adodb)
  ;;
esac

TAGS="${TAGS[@]}"
LDFLAGS=(
  -s
  -w
  -X github.com/xo/usql/text.CommandName=$NAME
  -X github.com/xo/usql/text.CommandVersion=$VER
)
LDFLAGS="${LDFLAGS[@]}"

(set -x;
  go build \
    -tags="$TAGS" \
    -ldflags="$LDFLAGS" \
    $@
)
