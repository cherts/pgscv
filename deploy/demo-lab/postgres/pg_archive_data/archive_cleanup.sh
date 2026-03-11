#!/bin/bash

PG_HOSTNAME=$(hostname)

if [ -d "/data/postgres_archive/${PG_HOSTNAME}" ]; then
    pg_archivecleanup -d /data/postgres_archive/${PG_HOSTNAME} $2
fi