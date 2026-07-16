#!/bin/sh

# Determine OS platform
# shellcheck source=/dev/null
. /etc/os-release

stop_pgscv_openrc() {
    echo "PreRemove: Stopping pgscv service"
    rc-service stop pgscv >/dev/null 2>&1 || true
}

case "$ID" in
    alpine)
        stop_pgscv_openrc
        ;;
esac

exit 0