#!/bin/sh
set -e
if ! getent group proxyglass >/dev/null; then
    groupadd -r proxyglass
fi
if ! getent passwd proxyglass >/dev/null; then
    useradd -r -g proxyglass -d /var/lib/proxyglass -s /sbin/nologin -c "Proxyglass" proxyglass
fi
