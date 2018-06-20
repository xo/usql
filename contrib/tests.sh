#!/bin/bash

rm -rf test.db

export USQL_SHOW_HOST_INFORMATION=false

autoexpect -f contrib/sqlite3/test.sql.exp ./usql file://test.db
