#!/bin/bash

TAGS="all fts5 vtable json1"

vgo build -tags "$TAGS" $@
