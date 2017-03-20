#!/bin/bash

# all supported databases
DATABASES="adodb avatica clickhouse couchbase odbc oracle ql saphana voltdb yql"

# additional sqlite3 tags
SQLITE3="icu fts5 vtable json1"

go build -tags "$DATABASES $SQLITE3" $@
