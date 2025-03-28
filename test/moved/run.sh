#!/usr/bin/env bash

set -eu

run() {
  rm yo_moved.tf
  tfmv -m yo_moved.tf --regexp '^test-/example-'
}

clean() {
  git checkout -- main.tf yo_moved.tf
}

run_test() {
  for file in main.tf yo_moved.tf; do
    if diff "$file" "${file}.after" >/dev/null; then
      echo "[ERROR] $file and ${file}.after is same before running tfmv" >&2
      return 1
    fi
  done
  
  run
  
  for file in main.tf yo_moved.tf; do
    if diff "$file" "${file}.after"; then
      git checkout -- "$file"
    else
      echo "[ERROR] $file and ${file}.after is different after running tfmv" >&2
      clean
      return 1
    fi
  done
  
  clean
}


case $1 in
  update)
    run
    for file in main.tf yo_moved.tf; do
      cp "$file" "${file}.after"
    done
    clean
    exit 0
    ;;
  test)
    run_test
    echo "[INFO] passed test" >&2
    exit 0
    ;;
  *)
    echo "[ERROR] The first argument must be either update or test" >&2
    exit 1
    ;;
esac
