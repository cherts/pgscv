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

# Check command exist function
_command_exists() {
	type "$1" &> /dev/null
}

# Detect Docker Compose
if _command_exists docker-compose; then
    DC_BIN=$(which docker-compose)
else
	echo "ERROR: docker-compose binary not found."
	exit 1
fi

echo "Stopping all container via docker-compose, please waiting..."
${DC_BIN} down >/dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "Remove grafana data..."
    rm -rf ${SCRIPT_DIR}/grafana/data/* >/dev/null 2>&1
    echo "Remove patroni data..."
    shopt -s dotglob
    rm -rf ${SCRIPT_DIR}/patroni/etc_data1/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/patroni/etc_data2/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/patroni/etc_data3/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/patroni/pg_data1/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/patroni/pg_data2/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/patroni/pg_data3/* >/dev/null 2>&1
    echo "Remove postgres data..."
    rm -rf ${SCRIPT_DIR}/postgres/pg12data/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/postgres/pg13data/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/postgres/pg14data/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/postgres/pg15data/* >/dev/null 2>&1
    rm -rf ${SCRIPT_DIR}/postgres/pg16data/* >/dev/null 2>&1
    echo "Remove victoriametrics data..."
    rm -rf ${SCRIPT_DIR}/victoriametrics/data/* >/dev/null 2>&1
    echo "All done."
else
    echo "ERROR: Container not stopped. Run 'docker-compose down' and see log."
    exit 1
fi
