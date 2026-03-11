#!/bin/bash

PG_HOSTNAME=$(hostname)

if [ ! -d "/data/postgres_archive/${PG_HOSTNAME}" ]; then
  mkdir -p "/data/postgres_archive/${PG_HOSTNAME}" 2>/dev/null
fi

gzip < $1 > /data/postgres_archive/${PG_HOSTNAME}/$2.gz