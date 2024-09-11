#!/bin/bash

# A single script to generate an entry in for userlist.txt
# Usage:
#
# ./generate_userlist.sh >> userlist.txt
# ./generate_userlist.sh username >> userlist.txt
#

# Don't edit this config
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
    DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
SCRIPT_NAME=$(basename "$0")

# Check command exist function
_command_exists() {
	type "$1" &> /dev/null
}

# Detect openssl
if _command_exists openssl; then
    OPENSSL_BIN=$(which openssl)
else
    echo "ERROR: Command 'openssl' not found."
    exit 1
fi

# Detect xxd
if _command_exists xxd; then
    XXD_BIN=$(which xxd)
else
    echo "ERROR: Command 'xxd' not found."
    exit 1
fi

if [[ $# -eq 1 ]]; then
  USERNAME="$1"
else
  read -r -p "Enter username: " USERNAME
fi

read -r -s -p "Enter password: " PASSWORD
echo >&2

# Using openssl md5 to avoid differences between OSX and Linux (`md5` vs `md5sum`)
ENC_PASSWORD="md5$(printf "%s%s" "${PASSWORD}" "${USERNAME}" | ${OPENSSL_BIN} md5 -binary | ${XXD_BIN} -p)"

echo "\"${USERNAME}\" \"${ENC_PASSWORD}\""