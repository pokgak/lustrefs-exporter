#!/bin/sh
set -e
systemctl daemon-reload
if [ "$1" = "configure" ] || [ "$1" = "1" ]; then
    systemctl enable lustrefs-exporter.service || true
fi
