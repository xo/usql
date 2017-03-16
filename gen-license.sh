#!/bin/bash

LICENSE=$(cat LICENSE)

DATA=$(cat << ENDSTR
package handler

const license = \`$LICENSE\`
ENDSTR
)

echo "$DATA" > handler/license.go
