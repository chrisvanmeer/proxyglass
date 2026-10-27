#!/bin/sh
set -e
if [ $1 -eq 0 ]; then
    systemctl stop proxyglass >/dev/null 2>&1 || true
    systemctl disable proxyglass >/dev/null 2>&1 || true
fi
