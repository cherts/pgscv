#!/bin/sh

# Determine OS platform
# shellcheck source=/dev/null
. /etc/os-release

stop_pgscv_systemd() {
    echo "Stopping pgscv service"
    systemctl stop pgscv >/dev/null 2>&1 || true
}

disable_pgscv_systemd() {
    echo "Disabling pgscv service"
    systemctl disable pgscv >/dev/null 2>&1 || true
}

systemd_daemon_reload() {
    echo "Running daemon-reload"
    systemctl daemon-reload || true
}

stop_pgscv_openrc() {
    echo "Stopping pgscv service"
    rc-service stop pgscv >/dev/null 2>&1 || true
}

disable_pgscv_openrc() {
    echo "Disabling pgscv service"
    rc-update del pgscv >/dev/null 2>&1 || true
}

case "$ID" in
    debian|ubuntu)
        case "$1" in
            remove|purge)
                stop_pgscv_systemd
                disable_pgscv_systemd
                systemd_daemon_reload
                ;;
        esac
        ;;
    rhel|fedora|centos|amzn|almalinux|rocky|ol)
        if [ "$1" = "0" ]; then
            # Package is being completely removed
            echo "PostRemove: Package being removed (not upgraded)"
            stop_pgscv_systemd
            disable_pgscv_systemd
            systemd_daemon_reload
        elif [ "$1" = "1" ]; then
            # Package is being upgraded
            echo "PostRemove: pgSCV is being upgraded, performing partial cleanup only"
            systemd_daemon_reload
        fi
        ;;
    alpine)
        stop_pgscv_openrc
        disable_pgscv_openrc
        ;;
    *)
        stop_pgscv_systemd
        disable_pgscv_systemd
        systemd_daemon_reload
        ;;
esac

exit 0