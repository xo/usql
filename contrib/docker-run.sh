#!/bin/bash

# docker-run.sh: starts or restarts docker containers.
#
# Usage: docker-run.sh <TARGET> [-u]
#
# Where <target> is a name of a subdirectory containing docker-config,
# 'all', or 'test'.
#
# all  -- starts all available database images.
# test -- starts the primary testing images. The testing images are cassandra,
#         mysql, postgres, sqlserver, and oracle [if available].
# -u   -- perform docker pull for images prior to start.
#
# Will stop any running docker container prior to starting.

DIR=$1

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}" )" && pwd))

if [ -z "$DIR" ]; then
  echo "usage: $0 <TARGET> [-u]"
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

  if [[ "$UPDATE" == "1" && "$TARGET" != "oracle" ]]; then
    if [ ! -f $BASE/Dockerfile ]; then
      (set -ex;
        docker pull $IMAGE
      )
    else
      pushd $BASE &> /dev/null
      (set -ex;
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

pushd $SRC &> /dev/null
TARGETS=()
case $DIR in
  all)
    TARGETS+=($(find . -type f -name docker-config|awk -F'/' '{print $2}'|grep -v oracle|grep -v db2))
    if [[ "$(docker image ls -q --filter 'reference=oracle/database')" != "" && -d /media/src/opt/oracle ]]; then
      TARGETS+=(oracle)
    fi
    if [[ "$(docker image ls -q --filter 'reference=ibmcom/db2')" != "" && -d /media/src/opt/db2 ]]; then
      TARGETS+=(db2)
    fi
  ;;
  test)
    TARGETS+=(mysql postgres sqlserver cassandra)
    if [[ "$(docker image ls -q --filter 'reference=oracle/database')" != "" && -d /media/src/opt/oracle ]]; then
      TARGETS+=(oracle)
    fi
  ;;
  *)
    TARGETS+=($DIR)
  ;;
esac

for TARGET in ${TARGETS[@]}; do
  docker_run $TARGET
done
popd &> /dev/null
