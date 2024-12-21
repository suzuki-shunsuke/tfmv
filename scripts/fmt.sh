#!/usr/bin/env bash

set -euo pipefail

git ls-files | grep -E "\.go$" | xargs gofumpt -w
