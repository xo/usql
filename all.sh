#!/bin/bash

# all supported databases
DATABASES="adodb avatica clickhouse couchbase firebird mymysql odbc oracle pgx ql saphana voltdb yql"

# additional sqlite3 tags
SQLITE3="icu fts5 vtable json1"

go build -tags "$DATABASES $SQLITE3" $@
