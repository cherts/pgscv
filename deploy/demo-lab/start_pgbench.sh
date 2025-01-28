#!/bin/bash

# postgres OR pgbouncer
PG_HOST="pgbouncer"
# default port
PG_PORT=5432
# base name of docker network
DOCKER_NETWORK="monitoring"

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

# Logging function
_logging() {
	local MSG=${1}
	local ENDLINE=${2:-"1"}
	if [[ "${ENDLINE}" -eq 0 ]]; then
		printf "%s: %s" "$(date "+%d.%m.%Y %H:%M:%S")" "${MSG}" 2>/dev/null
	else
		printf "%s: %s\n" "$(date "+%d.%m.%Y %H:%M:%S")" "${MSG}" 2>/dev/null
	fi
}

# Detect Docker
if _command_exists docker; then
	DOCKER_BIN=$(which docker)
else
	echo "ERROR: Command 'docker' not found."
	exit 1
fi

DATE_START=$(date +"%s")

PG_VERSIONS=(
	"9,1.4.5"
	"10,1.4.5"
	"11,1.4.5"
	"12,1.4.5"
	"13,1.4.6"
	"14,1.4.7"
	"15,1.4.8"
	"16,1.5.0"
	"17,1.5.0"
)

# Calculate duration function
_duration() {
	local DATE_START=${1:-"$(date +'%s')"}
	local FUNC_NAME=${2:-""}
	local DATE_END=$(date +"%s")
	local D_MSG=""
	local DATE_DIFF=$((${DATE_END} - ${DATE_START}))
	if [ -n "${FUNC_NAME}" ]; then
		local D_MSG=" of execute function '${FUNC_NAME}'"
	fi
	_logging "Duration${D_MSG}: $((${DATE_DIFF} / 3600)) hours $(((${DATE_DIFF} % 3600) / 60)) minutes $((${DATE_DIFF} % 60)) seconds"
}

_logging "Starting script."

for DATA in ${PG_VERSIONS[@]}; do
	PG_VER=$(echo "${DATA}" | awk -F',' '{print $1}')
	PGREPACK_VER=$(echo "${DATA}" | awk -F',' '{print $2}')
	_logging "Running pgbench for PostgreSQL v${PG_VER} in an infinite loop..."
	_logging "If you want to stop the test to create step-file '${SCRIPT_DIR}/pgbench/stop_pgbench_${PG_HOST}${PG_VER}_${PG_PORT}'"
	${DOCKER_BIN} run -it -d --rm --network "$(basename ${SCRIPT_DIR})_${DOCKER_NETWORK}" \
		--name pgbench_${PG_VER} \
		-e ${PWD}/pgbench/.env \
		-v ${PWD}/pgbench:/pg_repack \
		cherts/pg-repack:${PGREPACK_VER} bash -c "/pg_repack/start_pgbench_test.sh ${PG_VER} ${PG_HOST}${PG_VER} ${PG_PORT}" >/dev/null 2>&1
	if [ $? -eq 0 ]; then
		_logging "Done, container 'pgbench_${PG_VER}' is runned."
		_logging "View process: docker logs pgbench_${PG_VER} -f"
	else
		_logging "ERROR: Container 'pgbench_${PG_VER}' not runned."
	fi
done

_logging "Creating pgbench database for Patroni..."
${DOCKER_BIN} run -it --rm --network "$(basename ${SCRIPT_DIR})_${DOCKER_NETWORK}" \
	--name pgbench_patroni \
	-v ${PWD}/patroni/.env:/pg_repack/.env \
	-v ${PWD}/postgres/init.sql:/pg_repack/init.sql \
	cherts/pg-repack:1.5.0 bash -c "source /pg_repack/.env && psql -h haproxy -p 5000 -U postgres postgres -f /pg_repack/init.sql" >/dev/null 2>&1
if [ $? -eq 0 ]; then
	_logging "Done, database 'pgbench' is created."
else
	_logging "ERROR: Database 'pgbench' not created."
fi

_logging "Running pgbench for Patroni in an infinite loop..."
_logging "If you want to stop the test to create step-file '${SCRIPT_DIR}/pgbench/stop_pgbench_haproxy_5000'"
${DOCKER_BIN} run -it -d --rm --network "$(basename ${SCRIPT_DIR})_${DOCKER_NETWORK}" \
	--name pgbench_patroni \
	-v ${PWD}/pgbench:/pg_repack \
	cherts/pg-repack:1.5.0 bash -c "/pg_repack/start_pgbench_test.sh 16 haproxy 5000" >/dev/null 2>&1
if [ $? -eq 0 ]; then
	_logging "Done, container 'pgbench_patroni' is runned."
	_logging "View process: docker logs pgbench_patroni -f"
else
	_logging "ERROR: Container 'pgbench_patroni' not runned."
fi

_logging "Running pgbench (only select) for Patroni in an infinite loop..."
_logging "If you want to stop the test to create step-file '${SCRIPT_DIR}/pgbench/stop_pgbench_haproxy_5001'"
${DOCKER_BIN} run -it -d --rm --network "$(basename ${SCRIPT_DIR})_${DOCKER_NETWORK}" \
	--name pgbench_patroni_s \
	-v ${PWD}/pgbench:/pg_repack \
	cherts/pg-repack:1.5.0 bash -c "/pg_repack/start_pgbench_test.sh 16 haproxy 5001 1" >/dev/null 2>&1
if [ $? -eq 0 ]; then
	_logging "Done, container 'pgbench_patroni_s' is runned."
	_logging "View process: docker logs pgbench_patroni_s -f"
else
	_logging "ERROR: Container 'pgbench_patroni_s' not runned."
fi

_logging "All done."
_duration "${DATE_START}"

_logging "End script. Goodbye ;)"
