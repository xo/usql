#!/bin/bash

set -e

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../)

pushd $SRC &> /dev/null

#dep status -json|jq -r '.[]|select(.Constraint == "branch master" and .Version == "branch master" and (.Revision != .Latest))'
dep status -json|jq -r '.[]|select(.Version == "branch master" and (.Revision != .Latest))'

popd &> /dev/null
