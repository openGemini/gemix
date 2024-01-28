#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

exec env GODEBUG=madvdontneed=1 bin/ts-server \
    --config=conf/ts-server.toml \
    >> "{{.LogDir}}/server_extra.log"
