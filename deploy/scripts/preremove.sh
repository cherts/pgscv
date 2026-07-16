#!/bin/sh
set -e

# Stop the service before removal
systemctl stop pgscv || true
systemctl disable pgscv  || true

exit 0