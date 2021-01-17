#!/bin/bash

TAGS="all sqlite_app_armor sqlite_fts5 sqlite_icu sqlite_introspect sqlite_json1 sqlite_stat4 sqlite_userauth sqlite_vtable no_snowflake"

go build -tags "$TAGS" $@
