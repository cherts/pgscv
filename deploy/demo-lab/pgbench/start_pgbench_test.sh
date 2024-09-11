#!/bin/bash

PG_VER=$1
PG_HOST=$2
PG_PORT=$3

if [ -z "${PG_VER}" ]; then
    PG_VER=16
fi

if [ -z "${PG_HOST}" ]; then
    PG_HOST="postgres${PG_VER}"
fi

if [ -z "${PG_PORT}" ]; then
    PG_PORT=5432
fi

STOP_FLAG="/pg_repack/stop_pgbench_${PG_HOST}"
DATE_START=$(date +"%s")

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

# Calculate duration function
_duration() {
    local DATE_START=${1:-"$(date +'%s')"}
    local FUNC_NAME=${2:-""}
    local DATE_END=$(date +"%s")
    local D_MSG=""
    local DATE_DIFF=$((${DATE_END}-${DATE_START}))
    if [ -n "${FUNC_NAME}" ]; then
        local D_MSG=" of execute function '${FUNC_NAME}'"
    fi
    _logging "Duration${D_MSG}: $((${DATE_DIFF} / 3600 )) hours $(((${DATE_DIFF} % 3600) / 60)) minutes $((${DATE_DIFF} % 60)) seconds"
}

_logging "Starting script."

source /pg_repack/.env

_logging "Use pgbench for PostgreSQL v${PG_VER}, host=${PG_HOST}, port=${PG_PORT}"
_logging "STOP_FLAG: ${STOP_FLAG}"
rm -f "${STOP_FLAG}" >/dev/null 2>&1

_logging "Prepare pgbench database..."
pgbench -h ${PG_HOST} -p ${PG_PORT} -U pgbench pgbench -i -s 10

ITERATION=1
while true; do
    _logging "Run pgbench tests, iteration '${ITERATION}'..."
    pgbench -h ${PG_HOST} -p ${PG_PORT} -U pgbench pgbench -T 10 -j 4 -P 10 -c 5
    if [ -f "${STOP_FLAG}" ]; then
        _logging "Found stop-file '${STOP_FLAG}', end pgbench process."
        rm -f "${STOP_FLAG}" >/dev/null 2>&1
        break
    fi
    _logging "Run pgbench tests (select only), iteration '${ITERATION}'..."
    pgbench -h ${PG_HOST} -p ${PG_PORT} -U pgbench pgbench -T 10 -j 4 -P 10 -c 5 -S
    if [ -f "${STOP_FLAG}" ]; then
        _logging "Found stop-file '${STOP_FLAG}', end pgbench process."
        rm -f "${STOP_FLAG}" >/dev/null 2>&1
        break
    fi
    ((ITERATION++))
done

_logging "Remove pgbench database..."
pgbench -h ${PG_HOST} -p ${PG_PORT} -U pgbench pgbench -i -I d

_logging "All done."
_duration "${DATE_START}"

_logging "End script. Goodbye ;)"
