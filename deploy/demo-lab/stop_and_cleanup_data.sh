#!/bin/bash

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

# Detect Docker Compose
if _command_exists docker-compose; then
	DC_BIN=$(which docker-compose)
else
	echo "ERROR: docker-compose binary not found."
	exit 1
fi

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

echo "Stopping all container via docker-compose, please waiting..."
if [ -d "${SCRIPT_DIR}/victoriametrics/data/data" ]; then
	${DC_BIN} -f docker-compose.yml down --volumes >/dev/null 2>&1
else
	${DC_BIN} -f docker-compose.vm-cluster.yml down --volumes >/dev/null 2>&1
fi
if [ $? -eq 0 ]; then
	shopt -s dotglob
	echo "Remove grafana data..."
	rm -rf ${SCRIPT_DIR}/grafana/data/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/grafana/cluster_data/* >/dev/null 2>&1
	echo "Remove patroni data..."
	rm -rf ${SCRIPT_DIR}/patroni/etc_data1/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/patroni/etc_data2/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/patroni/etc_data3/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/patroni/pg_data1/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/patroni/pg_data2/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/patroni/pg_data3/* >/dev/null 2>&1
	for DATA in ${PG_VERSIONS[@]}; do
		PG_VER=$(echo "${DATA}" | awk -F',' '{print $1}')
		echo "Remove postgres v${PG_VER} data..."
		rm -rf ${SCRIPT_DIR}/postgres/pg${PG_VER}data/* >/dev/null 2>&1
		rm -rf ${SCRIPT_DIR}/postgres/pg${PG_VER}replica1data/* >/dev/null 2>&1
		rm -rf ${SCRIPT_DIR}/postgres/pg${PG_VER}replica2data/* >/dev/null 2>&1
	done
	echo "Remove victoriametrics data..."
	rm -rf ${SCRIPT_DIR}/victoriametrics/data/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/victoriametrics/vmstorage1data/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/victoriametrics/vmstorage2data/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/vmagent/data/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/vmagent/data1/* >/dev/null 2>&1
	rm -rf ${SCRIPT_DIR}/vmagent/data2/* >/dev/null 2>&1
	echo "All done."
else
	echo "ERROR: Container not stopped. Run 'docker-compose down' and see log."
	exit 1
fi
