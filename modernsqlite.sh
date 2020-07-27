#!/bin/bash

TAGS="no_base moderncsqlite"

go build -tags "$TAGS" $@
