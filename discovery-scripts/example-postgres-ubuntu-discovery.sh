#!/usr/bin/env bash
#
# Discovery PostgreSQL instance from Ubuntu via pg_lsclusters
#
# Usage: example-postgres-ubuntu-discovery.sh [PG_VERSION|ALL]
# PG_VERSION - 14/15/16/17/18/19/...
# ALL - discover all PostgreSQL instances
# Example: example-postgres-ubuntu-discovery.sh 18
#          example-postgres-ubuntu-discovery.sh ALL
#
# Author: Mikhail Grigorev

# User and password for PostgreSQL user (pgscv)
PG_USER="pgscv"
PG_PASSWORD="pgscv"

if [[ $# -eq 0 ]]; then
	ARGS="ALL"
else
	ARGS=${1:-"18"}
fi

# Check command exist function
_command_exists() {
	type "$1" &>/dev/null
}

if _command_exists pg_lsclusters; then
	PG_LSCLUSTERS_BIN=$(which pg_lsclusters)
else
	echo "ERROR: pg_lsclusters not found" >&2
	exit 2
fi

if _command_exists jq; then
	JQ_BIN=$(which jq)
else
	echo "ERROR: jq not found" >&2
	exit 2
fi

if _command_exists psql; then
	PSQL_BIN=$(which psql)
else
	echo "ERROR: psql not found" >&2
	exit 2
fi

if [[ "${ARGS}" == "ALL" ]]; then
	CLUSTER_LIST=$(${PG_LSCLUSTERS_BIN} -s -j 2>/dev/null | ${JQ_BIN} -r '.[].port')
else
	CLUSTER_LIST=$(${PG_LSCLUSTERS_BIN} -s -j 2>/dev/null | ${JQ_BIN} '.[] | select(.version == '\"${ARGS}\"')' | ${JQ_BIN} -r '.port')
fi

echo "# service-id host port database user password-from-env"

_discover_databases() {
	local PORT=${1:-"5432"}
	local SERVICE_ID=${2:-"svc-pg-${PORT}"}

	local DB_LIST=($(PGPASSWORD=${PG_PASSWORD} ${PSQL_BIN} -X -t -A -h "127.0.0.1" -p "${PORT}" -d "postgres" -U "${PG_USER}" -c "SELECT datname FROM pg_database WHERE datistemplate = false AND datname NOT IN ('postgres', 'template0', 'template1');" 2>/dev/null || true))

	if [ -z "${DB_LIST}" ]; then
		return
	fi

	while IFS= read -r DB; do
		[ -z "${DB}" ] && continue
		echo "${SERVICE_ID}-${DB}-${PORT} 127.0.0.1 ${PORT} ${DB} ${PG_USER}"
	done <<<"${DB_LIST}"
}

IFS=$'\n'
for PORT in ${CLUSTER_LIST[@]}; do
	CLUSTER_NAME=$(${PG_LSCLUSTERS_BIN} -s -j 2>/dev/null | ${JQ_BIN} '.[] | select(.port == '\"${PORT}\"')' | ${JQ_BIN} -r '.cluster')
	if [ -z "${CLUSTER_NAME}" ]; then
		continue
	fi
	#echo "svc-pg-${PORT} 127.0.0.1 ${PORT} postgres ${PG_USER} PG_PASSWORD"
	_discover_databases "${PORT}" "svc-pg-${CLUSTER_NAME}" | while IFS= read -r SCV; do
		[ -z "${SCV}" ] && continue
		echo "${SCV} PG_PASSWORD"
	done
done
