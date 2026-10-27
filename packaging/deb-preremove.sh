#!/bin/sh
set -e
if [ "$1" = "remove" ] || [ "$1" = "purge" ]; then
    systemctl stop proxyglass >/dev/null 2>&1 || true
    systemctl disable proxyglass >/dev/null 2>&1 || true
fi
