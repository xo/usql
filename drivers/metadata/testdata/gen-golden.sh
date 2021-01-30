#!/usr/bin/env bash

container=usql-pgsql

PGHOST="${PGHOST:-127.0.0.1}"
if [ -z "$PGPORT" ]; then
    port=$(docker port "$container" 5432/tcp)
    PGPORT=${port##*:}
fi
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-pw}"

export PGHOST PGPORT PGUSER PGPASSWORD

declare -A queries
queries=(
    [descTable]="\d film*"
    [listTables]="\dtvmsE film*"
)

for cmd in "${!queries[@]}"; do
    query="${queries[$cmd]}"
    psql --no-psqlrc --command "$query" --output "pgsql.$cmd.golden.txt"
done
