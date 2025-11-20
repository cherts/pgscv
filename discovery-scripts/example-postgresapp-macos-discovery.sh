#!/usr/bin/env bash
#
# Discovery PostgreSQL instance (use https://postgresapp.com) from macOS
#
# Usage: example-postgresapp-macos-discovery.sh [PG_VERSION|ALL]
# PG_VERSION - 14/15/16/17/18/...
# ALL - discover all PostgreSQL instances
# Example: example-postgresapp-macos-discovery.sh 15
#          example-postgresapp-macos-discovery.sh ALL
#
# Author: Mikhail Grigorev

if [[ $# -eq 0 ]]; then
	ARGS="ALL"
else
	ARGS=$1
fi

if [[ "${ARGS}" == "ALL" ]]; then
	PG_PROCESS_LIST=$(ps -A | grep postgres | grep "\-D")
else
	PG_PROCESS_LIST=$(ps -A | grep postgres | grep "\-D" | grep "var-${ARGS}")
fi

echo "# service-id host port database user password-from-env"

discover_databases() {
	local PORT=$1
	local SERVICE_ID=$2

	local DB_LIST=$(psql -X -p "${PORT}" -h 127.0.0.1 -t -A -c "
        SELECT datname
        FROM pg_database
        WHERE datistemplate = false
        AND datname NOT IN ('postgres', 'template0', 'template1')
    " postgres 2>/dev/null || true)

	if [ -z "${DB_LIST}}" ]; then
		return
	fi

	while IFS= read -r DB; do
		[ -z "${DB}" ] && continue
		echo "${SERVICE_ID}-${DB}-${PORT} 127.0.0.1 ${PORT} ${DB} pgscv"
	done <<<"${DB_LIST}"
}

IFS=$'\n'
for PG in ${PG_PROCESS_LIST[@]}; do
	PORT=$(echo ${PG} | sed -n 's/.*-p \([^ ]*\).*/\1/p')
	if [ -z "${PORT}" ]; then
		continue
	fi
	#echo "svc-pg-${PORT} 127.0.0.1 ${PORT} postgres pgscv PG_PASSWORD"
	discover_databases "${PORT}" "svc-pg" | while IFS= read -r SCV; do
		echo "${SCV} PG_PASSWORD"
	done
done
