#!/bin/bash

LICENSE=$(cat LICENSE)

DATA=$(cat << ENDSTR
package text

var License = \`$LICENSE\`
ENDSTR
)

echo "$DATA" > text/license.go
