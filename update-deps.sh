#!/bin/bash

# Updates the go module dependencies used by the build, then verifies that the
# result still builds.
#
# The module set is everything reachable from `go list -deps -test ./...` with
# every driver enabled, minus the modules listed in SKIP below. `go mod tidy`
# runs afterwards, and unless -V was passed the tree is vetted with every
# driver and test enabled and built with build.sh. On failure go.mod/go.sum are
# restored (override with -k) and the modules that changed are reported.
#
# Two ways to keep a dependency from being updated:
#
#   * add it to SKIP below -- it is not passed to `go get -u`, although another
#     module's requirements can still pull it forward
#   * `go mod edit -exclude=<module>@<version>` -- a hard block on a single
#     broken release, recorded in go.mod and honored by every go command
#
# usage: update-deps.sh [-n] [-k] [-V] [-x] [extra go get flags]
#
#   -n  dry run: report the available updates, change nothing
#   -k  keep go.mod/go.sum even when verification fails
#   -V  skip verification
#   -x  pass -v -x to go get

set -euo pipefail

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

# modules never passed to `go get -u`
SKIP=(
  # >= v1.4 dropped libcontainer/user, which ory/dockertest still imports
  github.com/opencontainers/runc
)

DRYRUN=0
KEEP=0
VERIFY=1
GETFLAGS=(-u)

usage() {
  sed -n '19,25p' "$0" | sed 's/^# \?//'
  exit 1
}

OPTIND=1
while getopts "nkVxh" opt; do
case "$opt" in
  n) DRYRUN=1 ;;
  k) KEEP=1 ;;
  V) VERIFY=0 ;;
  x) GETFLAGS+=(-v -x) ;;
  *) usage ;;
esac
done
shift $((OPTIND - 1))
GETFLAGS+=("$@")

cd $SRC

# BUILD_TAGS is exactly what build.sh uses, ALL_TAGS additionally enables every
# driver and any test-only code
BUILD_TAGS=$($SRC/build.sh -T)
ALL_TAGS="${BUILD_TAGS/most/all} test"

# updatable lists every non-main module providing a package in the build graph,
# minus SKIP
updatable() {
  go list -tags "$ALL_TAGS" -deps -test \
    -f '{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}' ./... \
    | grep . | sort -u \
    | comm -23 - <(printf '%s\n' ${SKIP[@]+"${SKIP[@]}"} | sort -u)
}

# outdated reports which of the passed modules have a newer version available
outdated() {
  go list -m -u -f '{{if .Update}}  {{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}' $@ \
    | grep . || echo "  (none)"
}

MODS=$(updatable)

echo "TAGS:        $ALL_TAGS"
echo "MODULES:     $(grep -c . <<< "$MODS")"
echo "SKIPPED:     ${SKIP[*]:-(none)}"

if [ "$DRYRUN" = "1" ]; then
  echo "AVAILABLE:"
  outdated $MODS
  exit
fi

BACKUP=$(mktemp -d)
trap 'rm -rf $BACKUP' EXIT
cp go.mod go.sum $BACKUP/

# restore reports the modules changed by this run and puts back the original
# go.mod/go.sum unless -k was passed
restore() {
  echo "CHANGED:"
  diff $BACKUP/go.mod $SRC/go.mod | grep -E '^[<>]' | sort -k2 || :
  if [ "$KEEP" = "1" ]; then
    echo "KEEPING:     go.mod/go.sum (-k)"
  else
    echo "RESTORING:   go.mod/go.sum"
    cp $BACKUP/go.mod $BACKUP/go.sum $SRC/
  fi
}

echo "UPDATING:"
if ! go get "${GETFLAGS[@]}" $MODS; then
  echo -e "\n\nERROR: go get failed -- add the offending module to SKIP in $0"
  restore
  exit 1
fi

(set -x;
  go mod tidy
)

if [ "$VERIFY" = "1" ]; then
  OK=1
  (set -x; go vet -tags "$ALL_TAGS" ./...) || OK=0
  if [ "$OK" = "1" ]; then
    (set -x; $SRC/build.sh -b) || OK=0
  fi
  if [ "$OK" = "0" ]; then
    echo -e "\n\nERROR: verification failed -- pin the offending module with 'go mod edit -exclude=' or add it to SKIP in $0"
    restore
    exit 1
  fi
fi

echo "REMAINING:"
outdated $MODS ${SKIP[@]+"${SKIP[@]}"}
