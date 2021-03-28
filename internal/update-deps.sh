#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../)

set -e

pushd $SRC &> /dev/null
(set -x;
  go get -u $@ $(go list -tags most -f '{{ join .Imports "\n" }}' ./internal/...)
)
PKGS=$(go list -tags most -f '{{ join .Imports "\n" }}'|grep 'github.com/xo/usql'|grep -v drivers|grep -v internal)
(set -x;
  go get -u $@ $PKGS
)
(set -x;
  go mod tidy
)
popd &> /dev/null
