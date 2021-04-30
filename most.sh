#!/bin/bash

NAME=usql
VER="$(date +%y.%m.%d)-dev"

TAGS=(
  most
  sqlite_icu
  sqlite_app_armor
  sqlite_fts5
  sqlite_introspect
  sqlite_json1
  sqlite_stat4
  sqlite_userauth
  sqlite_vtable
  osusergo
  netgo
  static_build
)
TAGS="${TAGS[@]}"

EXTLDFLAGS=(
  -fno-PIC
  -static
  -licuuc
  -licui18n
  -licudata
  -ldl
)
EXTLDFLAGS="${EXTLDFLAGS[@]}"

LDFLAGS=(
  -s
  -w
  -X github.com/xo/usql/text.CommandName=$NAME
  -X github.com/xo/usql/text.CommandVersion=$VER
  -linkmode=external
  -extldflags \'$EXTLDFLAGS\'
  -extld g++
)
LDFLAGS="${LDFLAGS[@]}"

(set -x;
  go build \
    -tags="$TAGS" \
    -gccgoflags="all=-DU_STATIC_IMPLEMENTATION" \
    -buildmode=pie \
    -ldflags="$LDFLAGS" \
    $@
)
