#!/bin/bash

# size.sh measures how much each driver adds to the size of the usql binary.
#
# It builds usql once with every driver, and once more for each driver with
# only that driver left out, through its no_<tag> build tag. The difference is
# what removing the driver saves. Two drivers that share a module each save
# only what the other does not also use, so the savings do not add up to the
# size of the binary. The charts feature is measured in the same way.
#
# The builds are stripped, as build.sh strips a release.

set -e

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))
OUT=

OPTIND=1
while getopts "o:h" opt; do
case "$opt" in
  o) OUT=$OPTARG ;;
  h)
    cat <<END
usage: size.sh [-o FILE]

Measures what each driver adds to a stripped build of usql.

  -o FILE   also write the table to FILE
END
    exit 0
  ;;
  *) exit 1 ;;
esac
done

DIR=$(mktemp -d)
trap 'rm -rf "$DIR"' EXIT

# the tags that build.sh uses for a release
TAGS=$("$SRC/build.sh" -T)

build() {
  (cd "$SRC" && go build -o "$DIR/usql" -trimpath -ldflags='-s -w' -tags "$1" .)
  stat -c %s "$DIR/usql"
}

# each driver and feature that gen.go writes a file for in internal
tags() {
  for f in "$SRC"/internal/*.go; do
    grep -q '^//go:build (!no_drivers ||' "$f" && basename "$f" .go
  done
}

FULL=$(build "$TAGS")

ROWS=()
for tag in $(tags); do
  size=$(build "$TAGS no_$tag")
  ROWS+=("$((FULL - size)) $tag")
done

{
  printf 'build with every driver: %.1f MiB, stripped, %s\n\n' "$(echo "$FULL/1048576" | bc -l)" "$(cd "$SRC" && go version)"
  printf '%10s  %s\n' 'MiB saved' 'driver left out'
  printf '%s\n' "${ROWS[@]}" | sort -rn | while read bytes tag; do
    printf '%10.1f  %s\n' "$(echo "$bytes/1048576" | bc -l)" "$tag"
  done
} | tee ${OUT:+"$OUT"}
