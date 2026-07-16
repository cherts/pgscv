#!/bin/sh

case "$1" in
    install)
        if ! getent passwd postgres > /dev/null 2>&1; then
            useradd --system --user-group postgres >/dev/null 2>&1 || true
        fi
        ;;
esac

exit 0