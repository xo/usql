#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}" )" && pwd))

for TARGET in $SRC/*/podman-config; do
  NAME=$(basename $(dirname $TARGET))
  if [ ! -z "$(podman ps -q --filter "name=$NAME")" ]; then
    (set -x;
      podman stop $NAME
    )
  fi
  if [ ! -z "$(podman ps -q -a --filter "name=$NAME")" ]; then
    (set -x;
      podman rm -f $NAME
    )
  fi
done
