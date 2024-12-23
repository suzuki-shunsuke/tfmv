#!/usr/bin/env bash

set -eu

f=$(mktemp)
cp main.tf "$f"
tfmv -r '-/_' --exclude '^github_'

if ! diff main.tf "$f"; then
  rm "$f"
  echo "[ERROR] main.tf is changed" >&2
  return 1
fi
if test -f moved.tf; then
  rm "$f"
  echo "[ERROR] moved.tf is created" >&2
  return 1
fi

rm "$f"
echo "[INFO] passed test" >&2
exit 0
