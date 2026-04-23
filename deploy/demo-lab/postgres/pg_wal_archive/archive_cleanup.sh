#!/bin/bash

if [ "$#" -ne 1 ]; then
    exit 1
fi

PG_HOSTNAME=$(hostname)

if [ -d "/data/pg_wal_archive/${PG_HOSTNAME}" ]; then
    pg_archivecleanup -d /data/pg_wal_archive/${PG_HOSTNAME} $1.gz
fi