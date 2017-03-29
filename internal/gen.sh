#!/bin/bash

BASE="mssql mysql postgres sqlite3"

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)

ALL=$(find $SRC/drivers/ -mindepth 1 -maxdepth 1 -type d|sort)

NL=$'\n'

# generate imports for all drivers
for i in $ALL; do
  MOST=""
  NAME=$(basename $i)

  TAGS="!no_$NAME"
  if ! [[ "$BASE" =~ "$NAME" && "$NAME" != "ql" ]]; then
    TAGS="all,!no_$NAME"

    if [[ "$NAME" != "odbc" && "$NAME" != "oracle" ]]; then
      TAGS="$TAGS most,!no_$NAME"
    fi

    TAGS="$TAGS $NAME,!no_$NAME"
  fi

  DATA=$(cat << ENDSTR
// +build $TAGS

package internal

import (
  // $NAME driver
  _ "github.com/knq/usql/drivers/$NAME"
)
ENDSTR
)
  echo "$DATA" > $NAME.go
  gofmt -w -s $NAME.go
done

KNOWN=
for i in $ALL; do
  NAME=$(basename $i)
  DRV=$(sed -n '/DRIVER: /p' $i/$NAME.go|sed -e 's/.*DRIVER:\s*//')
  PKG=$(sed -n '/DRIVER: /{n;p;}' $i/$NAME.go|sed -e 's/^\(\s\|"\|_\)\+//'|sed -e 's/"\s*//')
  KNOWN="$KNOWN$NL\"$DRV\": \"$NAME\", // $PKG"
done

DATA=$(cat << ENDSTR
// +build !no_base

package internal

//go:generate ./gen.sh

// KnownBuildTags returns a map of known driver names to its respective build
// tags.
func KnownBuildTags() map[string]string {
  return map[string]string{$KNOWN
  }
}
ENDSTR
)

echo "$DATA" > internal.go

gofmt -w -s internal.go
