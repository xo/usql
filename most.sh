#!/bin/bash

TAGS="most sqlite_app_armor sqlite_fts5 sqlite_icu sqlite_introspect sqlite_json1 sqlite_stat4 sqlite_userauth sqlite_vtable"

go build -tags "$TAGS" $@
