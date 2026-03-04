#!/usr/bin/env bash

PG_VER=${1:-"18"}
PGB_VERSION=${2:-"1.25.1"}
GO_VERSION="1.26.0"
REVIVE_VERSION="1.14.0"
GOSEC_VERSION="2.23.0"

# Don't edit this config
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
	DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
	SOURCE="$(readlink "$SOURCE")"
	[[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
SCRIPT_NAME=$(basename "$0")

# Check command exist function
_command_exists() {
	type "$1" &>/dev/null
}

_logging() {
    local MSG=${1}
    printf "%s: %s\n" "$(date "+%d.%m.%Y %H:%M:%S")" "${MSG}" 2>/dev/null
}

_logging "Use PostgreSQL v${PG_VER}"
_logging "Use PgBouncer v${PGB_VERSION}"

if [ ! -f "/usr/lib/postgresql/${PG_VER}/bin/initdb" ]; then
    _logging "PostgreSQL v${PG_VER} is not installed. Please install it first OR run this script with correct PostgreSQL version as argument."
    _logging "Example: ./${SCRIPT_NAME} <postgres_version> <pgbouncer_version>"
    exit 1
fi

_logging "Prepare OS dependencies..."
export PATH=$PATH:/usr/local/bin:/usr/local/go/bin
apt-get update
apt-get -y upgrade
apt-get -y install build-essential wget vim make gcc git curl libevent-dev libssl-dev pkg-config python3 pandoc

if ! _command_exists pgbouncer; then
    _logging "Set up PgBouncer ${PGB_VERSION}"
    wget https://www.pgbouncer.org/downloads/files/${PGB_VERSION}/pgbouncer-${PGB_VERSION}.tar.gz -O /tmp/pgbouncer-${PGB_VERSION}.tar.gz
    if [ -f "/tmp/pgbouncer-${PGB_VERSION}.tar.gz" ]; then
        tar -xzf /tmp/pgbouncer-${PGB_VERSION}.tar.gz -C /tmp
        cd /tmp/pgbouncer-${PGB_VERSION}
        ./configure --prefix=/usr/local
        make
        mkdir /etc/pgbouncer /var/log/pgbouncer /var/run/pgbouncer
        chown -R postgres:postgres /etc/pgbouncer /var/log/pgbouncer /var/run/pgbouncer
        cp pgbouncer /usr/sbin
        cp etc/pgbouncer.ini /etc/pgbouncer
        cd -
        rm -f /tmp/pgbouncer-${PGB_VERSION}.tar.gz
        rm -rf /tmp/pgbouncer-${PGB_VERSION}
    else
        _logging "ERROR: Failed to download PgBouncer v${PGB_VERSION}"
         exit 1
    fi
fi

if ! _command_exists go; then
    _logging "Set up Golang ${GO_VERSION}"
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm -f go${GO_VERSION}.linux-amd64.tar.gz
fi

if ! _command_exists revive; then
    _logging "Set up revive v${REVIVE_VERSION}"
    curl -s -L https://github.com/mgechev/revive/releases/download/v${REVIVE_VERSION}/revive_linux_amd64.tar.gz | tar xzf - -C $(go env GOROOT)/bin revive
fi

if ! _command_exists gosec; then
    _logging "Set up gosec v${GOSEC_VERSION}"
    curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOROOT)/bin v${GOSEC_VERSION}
fi

_logging "Prepare testing environment"
mkdir /opt/testing
chmod 755 /opt/testing
rm -rf /var/lib/apt/lists/*
git config --global --add safe.directory ${SCRIPT_DIR}/../

_logging "Show Golang version:"
go version
_logging "Show revive version:"
revive -version
_logging "Show gosec version:"
gosec -version
_logging "Show PgBouncer version:"
pgbouncer --version
_logging "Show PostgreSQL version:"
postgres --version

_logging "Copy test environment script"
cp ${SCRIPT_DIR}/prepare_test_environment.sh /opt/testing/
cp ${SCRIPT_DIR}/fixtures.sql /opt/testing/

_logging "Done."

_logging "Ready to prepare PostgreSQL and PgBouncer, please run manual command:"
echo "/opt/testing/prepare_test_environment.sh ${PG_VER}"
echo "export PATH=\$PATH:/usr/local/bin:/usr/local/go/bin"
echo "make test"
