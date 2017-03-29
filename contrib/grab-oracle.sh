#!/bin/bash

set -e

BASE=https://raw.githubusercontent.com/strongloop/loopback-oracle-builder/master/deps/oracle/Linux/x64

if [ ! -f instantclient-basiclite-linux.x64-12.1.0.2.0.zip ]; then
  wget $BASE/instantclient-basiclite-linux.x64-12.1.0.2.0.zip
fi

if [ ! -f instantclient-sdk-linux.x64-12.1.0.2.0.zip ]; then
  wget $BASE/instantclient-sdk-linux.x64-12.1.0.2.0.zip
fi

rm -rf instantclient_12_1

unzip -qq instantclient-basiclite-linux.x64-12.1.0.2.0.zip
unzip -qq instantclient-sdk-linux.x64-12.1.0.2.0.zip

pushd instantclient_12_1 &> /dev/null

ln -s libclntshcore.so.12.1 libclntshcore.so
ln -s libclntsh.so.12.1 libclntsh.so
ln -s libocci.so.12.1 libocci.so

popd &> /dev/null

DATA=$(cat << 'ENDSTR'
prefix=${pcfiledir}

version=12.1
build=client64

libdir=${prefix}/instantclient_12_1
includedir=${prefix}/instantclient_12_1/sdk/include

Name: OCI
Description: Oracle database engine
Version: ${version}
Libs: -L${libdir} -lclntsh
Libs.private:
Cflags: -I${includedir}
ENDSTR
)

echo "$DATA" > oci8.pc
