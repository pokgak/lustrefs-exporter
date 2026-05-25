#!/bin/sh
set -e
if systemctl is-active --quiet lustrefs-exporter.service; then
    systemctl stop lustrefs-exporter.service || true
fi
systemctl disable lustrefs-exporter.service || true
