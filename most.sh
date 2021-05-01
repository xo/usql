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
  netgo
)
TAGS="${TAGS[@]}"

EXTLDFLAGS=(
  -static
  $(pkg-config --libs icu-i18n)
  -lm
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
    -ldflags="$LDFLAGS" \
    $@
)
