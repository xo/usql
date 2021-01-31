#!/usr/bin/env bash

pgsql_in_docker=false
pgsql_container=usql-pgsql

if [ "$pgsql_in_docker" != true ]; then
    PGHOST="${PGHOST:-127.0.0.1}"
    port=$(docker port "$pgsql_container" 5432/tcp)
    PGPORT=${port##*:}
else
    PGHOST="${PGHOST:-$pgsql_container}"
    PGPORT=5432
fi
PGUSER="${PGUSER:-postgres}"
PGPASSWORD="${PGPASSWORD:-pw}"

export PGHOST PGPORT PGUSER PGPASSWORD

declare -A queries
queries=(
    [descTable]="\d film*"
    [listTables]="\dtvmsE film*"
)

for q in "${!queries[@]}"; do
    query="${queries[$q]}"
    cmd=(psql --no-psqlrc --command "$query")
    if [ "$pgsql_in_docker" == true ]; then
        docker run -it --rm -e PGHOST -e PGPORT -e PGUSER -e PGPASSWORD --link "$pgsql_container" postgres:13 "${cmd[@]}" >"pgsql.$q.golden.txt"
    else
        "${cmd[@]}" -o "pgsql.$q.golden.txt"
    fi
done

mysql_in_docker=true
mysql_container=usql-mysql

if [ "$mysql_in_docker" != true ]; then
    MYHOST="${MYHOST:-127.0.0.1}"
    port=$(docker port "$mysql_container" 3306/tcp)
    MYPORT=${port##*:}
else
    MYHOST="${MYHOST:-$mysql_container}"
    MYPORT=3306
fi
MYUSER="${MYUSER:-root}"
MYPASSWORD="${MYPASSWORD:-pw}"

declare -A queries
queries=(
    [descTable]="DESC film; DESC film_actor; DESC film_category; DESC film_list; DESC film_text;"
    [listTables]="SHOW TABLES LIKE 'film%'"
)

for q in "${!queries[@]}"; do
    query="${queries[$q]}"
    cmd=(mysql -h "$MYHOST" -P "$MYPORT" -u "$MYUSER" --password="$MYPASSWORD" --no-auto-rehash --database sakila --execute "$query")
    if [ "$mysql_in_docker" == true ]; then
        docker run -it --rm --link "$mysql_container" mysql:8 "${cmd[@]}" >"mysql.$q.golden.txt"
    else
        "${cmd[@]}" >"mysql.$q.golden.txt"
    fi
done
