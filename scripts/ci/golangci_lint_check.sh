#!/usr/bin/env bash

set -e

golangci-lint run --disable=staticcheck --timeout=10m --tests=false  ./...
