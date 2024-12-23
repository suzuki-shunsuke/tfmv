#!/usr/bin/env bash

set -euo pipefail


git checkout -- test # reset

while read -r dir; do
  if diff -r "$dir" "${dir/input/output}"
  pushd "$dir" > /dev/null

  # check *.tf and *.tf.after is same
  if diff 
  # check if moved.tf is empty
  # run command
  # check
  # check *.tf and *.tf.after is same

  popd > /dev/null
done < <(git ls-files test/input | grep test.sh | xargs -n 1 dirname)

git checkout -- test # reset
