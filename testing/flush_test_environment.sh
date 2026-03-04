#!/usr/bin/bash

PG_VER=${1:-"18"}
MAIN_DATADIR=/var/lib/postgresql/data/main
STDB1_DATADIR=/var/lib/postgresql/data/standby1
STDB2_DATADIR=/var/lib/postgresql/data/standby2
LGDB1_DATADIR=/var/lib/postgresql/data/logical1

# Don't edit this config
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
	DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
	SOURCE="$(readlink "$SOURCE")"
	[[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$(cd -P "$(dirname "$SOURCE")" && pwd)"
SCRIPT_NAME=$(basename "$0")

_logging() {
    local MSG=${1}
    printf "%s: %s\n" "$(date "+%d.%m.%Y %H:%M:%S")" "${MSG}" 2>/dev/null
}

_logging "Use PostgreSQL v${PG_VER}"

if [ ! -f "/usr/lib/postgresql/${PG_VER}/bin/initdb" ]; then
    _logging "PostgreSQL v${PG_VER} is not installed. Please install it first OR run this script with correct PostgreSQL version as argument."
    _logging "Example: ./${SCRIPT_NAME} <postgres_version>"
    exit 1
fi

_logging "Stop logical standby PostgreSQL v${PG_VER} via pg_ctl..."
su - postgres -c "/usr/lib/postgresql/${PG_VER}/bin/pg_ctl -D ${LGDB1_DATADIR} stop"

_logging "Remove logical standby data directory..."
shopt -s dotglob && rm -rf ${LGDB1_DATADIR}/*

_logging "Stop physical cascade standby PostgreSQL v${PG_VER} via pg_ctl..."
su - postgres -c "/usr/lib/postgresql/${PG_VER}/bin/pg_ctl -D ${STDB2_DATADIR} stop"

_logging "Remove physical cascade standby data directory..."
shopt -s dotglob && rm -rf ${STDB2_DATADIR}/*

_logging "Stop physical standby PostgreSQL v${PG_VER} via pg_ctl..."
su - postgres -c "/usr/lib/postgresql/${PG_VER}/bin/pg_ctl -D ${STDB1_DATADIR} stop"

_logging "Remove physical standby data directory..."
shopt -s dotglob && rm -rf ${STDB1_DATADIR}/*

_logging "Stop primary PostgreSQL v${PG_VER} via pg_ctl..."
su - postgres -c "/usr/lib/postgresql/${PG_VER}/bin/pg_ctl -D ${MAIN_DATADIR} stop"

_logging "Remove primary data directory..."
shopt -s dotglob && rm -rf ${MAIN_DATADIR}/*

_logging "Remove log file..."
rm -f /var/log/postgresql/*.log

_logging "Stop PgBouncer..."
su - postgres -c "kill -9 $(cat /var/run/pgbouncer/pgbouncer.pid)"
rm -f /var/run/pgbouncer/pgbouncer.pid
rm -f /var/log/pgbouncer/pgbouncer.log
rm -f /etc/pgbouncer/pgbouncer.ini
cp /etc/pgbouncer/pgbouncer.orig.ini /etc/pgbouncer/pgbouncer.ini
