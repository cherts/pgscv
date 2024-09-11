#!/bin/bash

# Don't edit this config
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do
    DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE"
done
SCRIPT_DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"
SCRIPT_NAME=$(basename "$0")

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

PG_VERSIONS=(
    "12,1.4.5"
    "13,1.4.6"
    "14,1.4.7"
    "15,1.4.8"
    "16,1.5.0"
)

_logging "Starting script."

for DATA in ${PG_VERSIONS[@]}; do
    PG_VER=$(echo "${DATA}" | awk -F',' '{print $1}')
    PGREPACK_VER=$(echo "${DATA}" | awk -F',' '{print $2}')
    STOP_FILE=${SCRIPT_DIR}/pgbench/stop_pgbench_${PG_VER}
    _logging "Creating stop-file '${STOP_FILE}'"
    touch "${STOP_FILE}" >/dev/null 2>&1
done

_logging "All done."
_duration "${DATE_START}"

_logging "End script. Goodbye ;)"