#!/bin/bash

if [ "$#" -ne 2 ]; then
    exit 1
fi

PG_HOSTNAME=$(hostname)

gzip < /data/postgres/$1 > /data/pg_wal_archive/${PG_HOSTNAME}/$2.gz