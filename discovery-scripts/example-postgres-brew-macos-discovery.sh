#!/usr/bin/env bash
#
# Discovery PostgreSQL instance (use brew) from macOS
#
# Usage: example-postgres-brew-macos-discovery.sh
#
# Author: Dmitry Bulashev

set -o errexit
set -o pipefail
set -o nounset

echo "# service-id host port database user password-from-env"

discover_homebrew_postgres() {
	#echo "Checking Homebrew PostgreSQL services..." >&2

	if ! command -v brew >/dev/null 2>&1; then
		#echo "Homebrew not found" >&2
		return
	fi

	brew services list | while read -r service_info; do
		service_name=$(echo "$service_info" | awk '{print $1}')
		status=$(echo "$service_info" | awk '{print $2}')

		if [[ "$service_name" == *postgres* ]] && [ "$status" = "started" ]; then
			#echo "Found running PostgreSQL service: $service_name" >&2

			port=$(get_homebrew_service_port "$service_name")
			if [ -n "$port" ]; then
				service_id="macos_${service_name}_${port}"
				discover_databases "$port" "$service_id"
			fi
		fi
	done
}

get_homebrew_service_port() {
	local service_name=$1
	echo $service_name >&2
	local paths=(
		"/opt/homebrew/var/${service_name}"
	)

	for path in "${paths[@]}"; do
		local config_file="${path}/postgresql.conf"
		if [ -f "$config_file" ]; then
			local port=$(grep -E "^port\s*=" "$config_file" | awk -F= '{print $2}' | tr -d ' ' | head -1)
			if [ -n "$port" ]; then
				echo "$port"
				return
			fi
		fi
	done

	echo "5432"
}

discover_databases() {
	local port=$1
	local service_id_prefix=$2

	#echo "Discovering databases on port $port..." >&2

	local databases=$(psql -X -p "$port" -h localhost -t -A -c "
        SELECT datname
        FROM pg_database
        WHERE datistemplate = false
        AND datname NOT IN ('postgres', 'template0', 'template1')
    " postgres 2>/dev/null || true)

	if [ -z "$databases" ]; then
		#echo "Could not connect to PostgreSQL on port $port or not found any databases" >&2
		return
	fi

	while IFS= read -r dbname; do
		[ -z "$dbname" ] && continue
		service_id="${service_id_prefix}_${dbname}"
		echo "$service_id 127.0.0.1 $port $dbname pgscv PG_PASSWORD"
	done <<<"$databases"
}

main() {
	#echo "macOS PostgreSQL Discovery started on $(hostname)" >&2
	discover_homebrew_postgres
	#echo "Discovery completed" >&2
}

main

exit 0
