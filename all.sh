#!/bin/bash

TAGS="all icu fts5 vtable json1"

go build -tags "$TAGS" $@
