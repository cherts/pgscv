#!/bin/bash

PG_HOSTNAME=$(hostname)

if [ ! -d "/data/pg_wal_archive/${PG_HOSTNAME}" ]; then
  mkdir -p "/data/pg_wal_archive/${PG_HOSTNAME}" 2>/dev/null
  chown -R postgres:postgres "/data/pg_wal_archive/${PG_HOSTNAME}" 2>/dev/null
fi