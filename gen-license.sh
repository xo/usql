#!/bin/bash

LICENSE=$(cat LICENSE)

DATA=$(cat << ENDSTR
package text

// License contains the license text for usql.
var License = \`$LICENSE\`
ENDSTR
)

echo "$DATA" > text/license.go
