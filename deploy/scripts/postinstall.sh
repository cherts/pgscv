#!/bin/sh

restart_pgscv_if_required() {
    if service pgscv status >/dev/null 2>&1; then
        printf "PostInstall: Restarting pgscv service\n"
        service pgscv restart >/dev/null 2>&1 || true
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

{
    restart_pgscv_if_required
    summary
}

exit 0