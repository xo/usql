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

source $BASE/docker-config

if [[ "$DIR" != "$NAME" ]]; then
  echo "error: $BASE/docker-config is invalid"
  exit 1
fi

echo "NAME:       $NAME"

shift

UPDATE=0

OPTIND=1
while getopts "dfuv:" opt; do
case "$opt" in
  u) UPDATE=1 ;;
esac
done

# setup params
declare -a PARAMS
for k in NAME PUBLISH ENV VOLUME NETWORK PRIVILEGED; do
  n=$(tr 'A-Z' 'a-z' <<< "$k")
  v=$(eval echo "\$$k")
  if [ ! -z "$v" ]; then
    for p in $v; do
      PARAMS=("${PARAMS[@]}" "--$n=$p")
    done
  fi
done

echo "IMAGE:      $IMAGE (update: $UPDATE)"
echo "PUBLISH:    $PUBLISH"
echo "ENV:        $ENV"
echo "VOLUME:     $VOLUME"
echo "NETWORK:    $NETWORK"
echo "PRIVILEGED: $PRIVILEGED"

if [ "$UPDATE" -eq "1" ]; then
  if [ ! -f $BASE/Dockerfile ]; then
    docker pull $IMAGE
  else
    pushd $BASE &> /dev/null
    docker build --pull -t $IMAGE:latest .
    popd &> /dev/null
  fi
fi

if [ ! -z "$(docker ps -q --filter "name=$NAME")" ]; then
  docker stop $NAME
fi

if [ ! -z "$(docker ps -q -a --filter "name=$NAME")" ]; then
  docker rm -f $NAME
fi

set -e

echo docker run --detach --rm ${PARAMS[@]} $IMAGE $EXTRA
     docker run --detach --rm ${PARAMS[@]} $IMAGE $EXTRA
