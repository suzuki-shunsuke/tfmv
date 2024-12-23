#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")"

for dir in arg dry-run exclude-1 exclude-2 include-1 include-2 jsonnet moved recursive regexp replace; do
  pushd "$dir"
  bash run.sh test
  popd
done
