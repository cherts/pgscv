#!/bin/sh

# Determine OS platform
# shellcheck source=/dev/null
. /etc/os-release

restart_pgscv_if_required() {
    if service pgscv status >/dev/null 2>&1; then
        echo "PostInstall: Restarting pgscv service"
        service pgscv restart >/dev/null 2>&1 || true
    fi
}

add_pgscv_openrc() {
    echo "PostInstall: Adding pgscv to openrc"
    rc-update add pgscv default >/dev/null 2>&1 || true
}

set_config_owner() {
    if getent passwd postgres > /dev/null 2>&1; then
        echo "PostInstall: Setting owner for configuration file"
        chown postgres:postgres /etc/pgscv.yaml >/dev/null 2>&1 || true
    fi
}

summary() {
    echo "----------------------------------------------------------------------"
    echo " pgSCV package has been successfully installed."
    echo ""
    echo " Please follow the next steps to start the software:"
    echo "    sudo systemctl enable pgscv --now"
    echo ""
    echo " Configuration settings can be adjusted here:"
    echo "    /etc/pgscv.yaml"
    echo ""
    echo "----------------------------------------------------------------------"
}

summary_alpine() {
    echo "----------------------------------------------------------------------"
    echo " pgSCV package has been successfully installed."
    echo ""
    echo " Please follow the next steps to start the software:"
    echo "    rc-service pgscv start"
    echo ""
    echo " Configuration settings can be adjusted here:"
    echo "    /etc/pgscv.yaml"
    echo ""
    echo "----------------------------------------------------------------------"
}

set_config_owner
restart_pgscv_if_required
case "$ID" in
    alpine)
        add_pgscv_openrc
        summary_alpine
        ;;
    *)
        summary
        ;;
esac

exit 0