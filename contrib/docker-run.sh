#!/bin/bash

DIR=$1

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}" )" && pwd))

if [ -z "$DIR" ]; then
  echo "usage: $0 <NAME>"
  exit 1
fi

shift

UPDATE=0

OPTIND=1
while getopts "u" opt; do
case "$opt" in
  u) UPDATE=1 ;;
esac
done

docker_run() {
  TARGET=$1
  BASE=$SRC/$TARGET
  if [ ! -e $BASE/docker-config ]; then
    echo "error: $BASE/docker-config doesn't exist"
    exit 1
  fi
  unset IMAGE NAME PUBLISH ENV VOLUME NETWORK PRIVILEGED PARAMS
  source $BASE/docker-config
  if [[ "$TARGET" != "$NAME" ]]; then
    echo "error: $BASE/docker-config is invalid"
    exit 1
  fi
  echo "-------------------------------------------"
  echo "NAME:       $NAME"
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

  EXISTS=$(docker image ls -q $IMAGE)
  if [[ "$UPDATE" == "0" && -z "$EXISTS" ]]; then
    UPDATE=1
  fi
  echo "IMAGE:      $IMAGE (update: $UPDATE)"
  echo "PUBLISH:    $PUBLISH"
  echo "ENV:        $ENV"
  echo "VOLUME:     $VOLUME"
  echo "NETWORK:    $NETWORK"
  echo "PRIVILEGED: $PRIVILEGED"

  if [ "$UPDATE" -eq "1" ]; then
    if [ ! -f $BASE/Dockerfile ]; then
      (set -x;
        docker pull $IMAGE
      )
    else
      pushd $BASE &> /dev/null
      (set -x;
        docker build --pull -t $IMAGE:latest .
      )
      popd &> /dev/null
    fi
  fi
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
  (set -ex;
    docker run --detach --rm ${PARAMS[@]} $IMAGE
  )
}

if [ "$DIR" = "test" ]; then
  for TARGET in mysql postgres sqlserver cassandra; do
    docker_run $TARGET
  done
else
  docker_run $DIR
fi
