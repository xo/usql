#!/bin/bash

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}" )" && pwd))

for TARGET in $SRC/*/docker-config; do
  NAME=$(basename $(dirname $TARGET))
  if [ ! -z "$(docker ps -q --filter "name=$NAME")" ]; then
    (set -x;
      docker stop $NAME
    )
  fi
  if [ ! -z "$(docker ps -q -a --filter "name=$NAME")" ]; then
    (set -x;
      docker rm -f $NAME
    )
  fi
done
