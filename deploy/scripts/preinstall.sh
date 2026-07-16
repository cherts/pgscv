#!/bin/sh
set -e

# $1 = "install" or "upgrade"
if [ "$1" = "install" ]; then
    echo "Executing actions before clean installation..."
    systemctl stop pgscv || true
    if ! getent passwd postgres > /dev/null 2>&1; then
        useradd --system --user-group postgres
    fi
elif [ "$1" = "upgrade" ]; then
    echo "Executing actions before upgrading from version $2..."
    systemctl stop pgscv || true
fi

exit 0