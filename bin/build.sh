#!/bin/bash
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}")" && pwd )
ROOT=$DIR/..

pushd $ROOT/src/gentable >/dev/null 2>&1 || exit
echo "Build gentable"
go build -o $ROOT/bin/gentable .
popd >/dev/null 2>&1 || exit