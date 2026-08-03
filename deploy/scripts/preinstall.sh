#!/bin/sh

# Determine OS platform
# shellcheck source=/dev/null
. /etc/os-release

add_user_and_group() {
    if ! getent passwd postgres > /dev/null 2>&1; then
        echo "PreInstall: Creating postgres user"
        useradd --system --user-group postgres >/dev/null 2>&1 || true
    fi
}

case "$ID" in
    debian|ubuntu)
        case "$1" in
            install)
                add_user_and_group
            ;;
        esac
        ;;
    rhel|fedora|centos|amzn|almalinux|rocky|ol)
        if [ "$1" = "1" ]; then
            # Package is being completely installed
            add_user_and_group
        fi
        ;;
    alpine)
        add_user_and_group
        ;;
    *)
        add_user_and_group
        ;;
esac

exit 0