#!/bin/bash
# Run every command of commands.txt on psql, the old usql and the new usql, as
# csv, into out/. Needs a PostgreSQL that holds fixture.sql in a database, and
# these variables:
#
#   PGURL     the URL of that database
#   USQL_OLD  the usql of the last release, such as v0.21.6
#   USQL_NEW  the usql to check
#
# Then run compare.py, values.py and sections.py, which read out/.
set -e
D=$(cd "$(dirname "$0")" && pwd)
: "${PGURL:?}" "${USQL_OLD:?}" "${USQL_NEW:?}"
mkdir -p $D/out/psql $D/out/old $D/out/new
n=0
while IFS= read -r c; do
  n=$((n+1)); f=$(printf '%03d' $n)
  echo "$c" > $D/out/cmd.$f
  timeout 30 psql "$PGURL" -X -P format=csv -P null='<NULL>' -c "$c" > $D/out/psql/$f.csv 2>&1 < /dev/null || true
  timeout 30 $USQL_OLD "$PGURL" -c '\pset format csv' -c '\pset null <NULL>' -c "$c" > $D/out/old/$f.csv 2>&1 < /dev/null || true
  timeout 30 $USQL_NEW "$PGURL" -c '\pset format csv' -c '\pset null <NULL>' -c "$c" > $D/out/new/$f.csv 2>&1 < /dev/null || true
done < $D/commands.txt
echo ran $n
