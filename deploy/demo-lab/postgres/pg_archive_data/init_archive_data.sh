#!/bin/bash

PG_HOSTNAME=$(hostname)

if [ ! -d "/data/postgres_archive/${PG_HOSTNAME}" ]; then
  mkdir -p "/data/postgres_archive/${PG_HOSTNAME}" 2>/dev/null
  chown -R postgres:postgres "/data/postgres_archive/${PG_HOSTNAME}" 2>/dev/null
fi