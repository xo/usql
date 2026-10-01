#!/bin/bash

# size.sh measures how much each driver adds to the size of the usql binary.
#
# It builds usql once with every driver, and once more for each driver with
# only that driver left out. The difference is what removing the driver saves.
# Two drivers that share a module each save only what the other does not also
# use, so the savings do not add up to the size of the binary.
#
# The builds are stripped, as build.sh strips a release.

set -e

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))
GROUP=all
OUT=

OPTIND=1
while getopts "g:o:h" opt; do
case "$opt" in
  g) GROUP=$OPTARG ;;
  o) OUT=$OPTARG ;;
  h)
    cat <<END
usage: size.sh [-g GROUP] [-o FILE]

Measures what each driver adds to a stripped build of usql.

  -g GROUP  the build group to measure: base, most or all (default: all)
  -o FILE   also write the table to FILE
END
    exit 0
  ;;
  *) exit 1 ;;
esac
done

DIR=$(mktemp -d)
trap 'rm -rf "$DIR"' EXIT

# the sqlite tags that build.sh adds to a release
TAGS=$("$SRC/build.sh" -T | tr ' ' '\n' | grep '^sqlite_' | tr '\n' ' ')

build() {
  (cd "$SRC" && go build -o "$DIR/usql" -trimpath -ldflags='-s -w' -tags "$1" .)
  stat -c %s "$DIR/usql"
}

# each driver package whose build line names the group
drivers() {
  for f in "$SRC"/internal/*.go; do
    tag=$(basename "$f" .go)
    case $tag in
      internal|z|charts) continue ;;
    esac
    case $GROUP in
      all) grep -q '^//go:build (\(!no_base\|all\)' "$f" && echo "$tag" ;;
      most) grep -q '^//go:build (\(!no_base\|all || most\)' "$f" && echo "$tag" ;;
      base) grep -q '^//go:build (!no_base' "$f" && echo "$tag" ;;
    esac
  done
}

FULL=$(build "$GROUP $TAGS")

ROWS=()
for tag in $(drivers); do
  size=$(build "$GROUP $TAGS no_$tag")
  ROWS+=("$((FULL - size)) $tag")
done

{
  printf '%s build: %.1f MB, stripped, %s\n\n' "$GROUP" "$(echo "$FULL/1048576" | bc -l)" "$(cd "$SRC" && go version)"
  printf '%10s  %s\n' 'MB saved' 'driver left out'
  printf '%s\n' "${ROWS[@]}" | sort -rn | while read bytes tag; do
    printf '%10.1f  %s\n' "$(echo "$bytes/1048576" | bc -l)" "$tag"
  done
} | tee ${OUT:+"$OUT"}
