#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

set -e

pushd $SRC &> /dev/null
(set -x;
  go get -u -v -x $@ $(go list -tags 'all test' -f '{{ join .Imports "\n" }}' ./internal/...)
)
PKGS=$(go list -tags 'all test' -f '{{ join .Imports "\n" }}'|grep 'github.com/xo/usql'|grep -v drivers|grep -v internal)
(set -x;
  go get -u -v -x $@ $PKGS
)
(set -x;
  go mod tidy
)
popd &> /dev/null
