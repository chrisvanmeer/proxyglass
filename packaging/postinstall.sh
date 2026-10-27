#!/bin/sh
set -e
chown -R root:proxyglass /etc/proxyglass 2>/dev/null || true
chmod 750 /etc/proxyglass 2>/dev/null || true
[ -f /etc/proxyglass/config.yml ] && chmod 640 /etc/proxyglass/config.yml && chown root:proxyglass /etc/proxyglass/config.yml || true
chown -R proxyglass:proxyglass /var/lib/proxyglass /var/log/proxyglass 2>/dev/null || true
systemctl daemon-reload >/dev/null 2>&1 || true
