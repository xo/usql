#!/bin/bash

DIR=$1

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

BASE=$SRC/$DIR

if [ -z "$DIR" ]; then
  echo "usage: $0 <NAME>"
  exit 1
fi

if [ ! -e $BASE/docker-config ]; then
  echo "error: $BASE/docker-config doesn't exist"
  exit 1
fi

. $BASE/docker-config

if [[ "$DIR" != "$NAME" ]]; then
  echo "error: $BASE/docker-config is invalid"
  exit 1
fi

echo "NAME: $NAME"

shift

# setup params
declare -a PARAMS
for k in NAME PUBLISH NETWORK VOLUME ENV; do
  n=$(tr 'A-Z' 'a-z' <<< "$k")
  v=$(eval echo "\$$k")
  if [ ! -z "$v" ]; then
    for p in $v; do
      PARAMS=("${PARAMS[@]}" "--$n=$p")
    done
  fi
done

docker stop $NAME

docker rm $NAME

set -e

echo docker run --detach --rm ${PARAMS[@]} $IMAGE
     docker run --detach --rm ${PARAMS[@]} $IMAGE
