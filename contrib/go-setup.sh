#!/bin/bash

# trimmed down version of:
# https://github.com/kenshaw/shell-config/blob/master/scripts/go-setup.sh

ARCH=$(uname -m)
PLATFORM=linux

case $ARCH in
  aarch64) ARCH=arm64 ;;
  x86_64)  ARCH=amd64 ;;
esac

REPO=https://go.googlesource.com/go
DL=https://go.dev/dl/
EXT=tar.gz

DEST=/usr/local

set -e

LATEST=$(curl -4 -s "$DL"|sed -E -n "/<a .+?>go1\.[0-9]+(\.[0-9]+)?\.$PLATFORM-$ARCH\.$EXT</p"|head -1)
ARCHIVE=$(sed -E -e 's/.*<a .+?>(.+?)<\/a.*/\1/' <<< "$LATEST")
STABLE=$(sed -E -e 's/^go//' -e "s/\.$PLATFORM-$ARCH\.$EXT$//" <<< "$ARCHIVE")

if ! [[ "$STABLE" =~ ^1\.[0-9\.]+$ ]]; then
  echo "ERROR: unable to retrieve latest Go version for $PLATFORM/$ARCH ($STABLE)"
  exit 1
fi

REMOTE=$(sed -E -e 's/.*<a .+?href="(.+?)".*/\1/' <<< "$LATEST")
VERSION="go$STABLE"

OPTIND=1
while getopts "v:" opt; do
case "$opt" in
  v) VERSION=$OPTARG ;;
esac
done

# prefix passed version with go
if [[ "$VERSION" =~ ^1\.[0-9]+ ]]; then
  VERSION="go$VERSION"
fi

if ! [[ "$VERSION" =~ ^go1\.[0-9]+\.[0-9]+$ ]]; then
  echo "ERROR: invalid Go version $VERSION"
  exit 1
fi

if ! [[ "$REMOTE" =~ "^https://" ]]; then
  REMOTE="https://go.dev$REMOTE"
fi

echo "ARCH:       $PLATFORM/$ARCH"
echo "DEST:       $DEST"
echo "STABLE:     $STABLE ($REMOTE)"
echo "VERSION:    $VERSION"

grab() {
  echo "RETRIEVING: $1 -> $2"
  curl -4 -L -# -o $2 $1
}

# extract
WORKDIR=$(mktemp -d /tmp/go-setup.XXXX)
grab $REMOTE $WORKDIR/$ARCHIVE
echo "USING:      $WORKDIR/$ARCHIVE"

pushd $WORKDIR &> /dev/null
case $EXT in
  tar.gz) tar -zxf $ARCHIVE ;;
  zip)    unzip -q $ARCHIVE ;;
  *)
    echo "ERROR:      unknown extension $EXT"
    exit
  ;;
esac

echo "MOVING:     $WORKDIR/go -> $DEST/go"
mv go $DEST/go

chown -R root:root $DEST/go

echo "INSTALLED:  $($DEST/go/bin/go version)"
