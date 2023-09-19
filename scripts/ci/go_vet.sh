#!/usr/bin/env bash

set -e

go list ./... | xargs go vet --tests=false