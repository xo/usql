#!/bin/bash

TAGS="all fts5 vtable json1 no_ql"

vgo build -tags "$TAGS" $@
