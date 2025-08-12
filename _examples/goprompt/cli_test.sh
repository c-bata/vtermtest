#!/bin/sh

set -eux

go run ../../cmd/vtermtest-cli/main.go \
  --command "go run simple_example/main.go" \
  --rows 10 --cols 80 \
  --delimiter "[]" --keys "[WaitFor >>>]u[Tab][WaitFor users]"

