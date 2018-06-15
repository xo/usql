#!/bin/bash

../source/runConfigureICU \
  MinGW \
  --host=x86_64-w64-mingw32 \
  --disable-release \
  --disable-debug \
  --enable-static \
  --prefix=/opt/local
