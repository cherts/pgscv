#!/bin/sh
set -e

# $1 = "install" or "upgrade"
if [ "$1" = "install" ]; then
    echo "Executing actions after clean installation..."
    systemctl enable pgscv --now || true
elif [ "$1" = "upgrade" ]; then
    echo "Executing actions after upgrading from version $2..."
    systemctl start pgscv || true
fi

exit 0