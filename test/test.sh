#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")"

while read file; do
  dir=$(dirname "$file")
  pushd "$dir"
  bash run.sh test
  popd
done < <(find . -name run.sh)
