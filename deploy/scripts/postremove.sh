#!/bin/sh

# Determine OS platform
# shellcheck source=/dev/null
. /etc/os-release

stop_pgscv_systemd() {
    echo "PostRemove: Stopping pgscv service"
    systemctl stop pgscv >/dev/null 2>&1 || true
}

disable_pgscv_systemd() {
    echo "PostRemove: Disabling pgscv service"
    systemctl disable pgscv >/dev/null 2>&1 || true
}

systemd_daemon_reload() {
    echo "PostRemove: Running daemon-reload"
    systemctl daemon-reload || true
}

restart_pgscv_if_required_systemd() {
    if service pgscv status >/dev/null 2>&1; then
        echo "PostRemove: Restarting pgscv service"
        systemctl restart pgscv >/dev/null 2>&1 || true
    fi
}

stop_pgscv_if_required_openrc() {
    if service pgscv status >/dev/null 2>&1; then
        echo "PostRemove: Stopping pgscv service"
        service pgscv stop >/dev/null 2>&1 || true
    fi
}

disable_pgscv_openrc() {
    echo "PostRemove: Disabling pgscv service"
    rc-update del pgscv default >/dev/null 2>&1 || true
}

cleanup_pgscv_config() {
    echo "PostRemove: Cleaning up pgscv configuration file"
    rm -f /etc/pgscv.yaml >/dev/null 2>&1 || true
}

cleanup_pgscv_pid() {
    echo "PostRemove: Cleaning up pgscv PID file"
    rm -f /run/pgscv.pid >/dev/null 2>&1 || true
}

case "$ID" in
    debian|ubuntu)
        case "$1" in
            remove)
                stop_pgscv_systemd
                disable_pgscv_systemd
                systemd_daemon_reload
                ;;
            purge)
                cleanup_pgscv_config
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
            echo "PostRemove: pgSCV is being upgraded, performing partial cleanup and restart only"
            systemd_daemon_reload
            restart_pgscv_if_required_systemd
        fi
        ;;
    alpine)
        stop_pgscv_if_required_openrc
        disable_pgscv_openrc
        cleanup_pgscv_config
        cleanup_pgscv_pid
        ;;
    *)
        stop_pgscv_systemd
        disable_pgscv_systemd
        systemd_daemon_reload
        cleanup_pgscv_config
        ;;
esac

exit 0