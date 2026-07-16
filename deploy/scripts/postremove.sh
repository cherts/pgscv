#!/bin/sh
set -e

# Reload systemd configuration after removal
systemctl daemon-reload || true

exit 0