#!/bin/bash

TAGS="all fts5 vtable json1"

go build -tags "$TAGS" $@
