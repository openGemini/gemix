#!/bin/bash
set -e

# WARNING: This file was auto-generated. Do not edit!
#          All your edit might be overwritten!
DEPLOY_DIR={{.DeployDir}}

cd "${DEPLOY_DIR}" || exit 1

exec env GODEBUG=madvdontneed=1 bin/ts-monitor \
    --config=conf/ts-monitor.toml \
    >> "{{.LogDir}}/monitor_extra.log"
