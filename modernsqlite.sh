#!/bin/bash

TAGS="most no_sqlite3 moderncsqlite"

CGO_ENABLED=0 go build -tags "$TAGS" $@
