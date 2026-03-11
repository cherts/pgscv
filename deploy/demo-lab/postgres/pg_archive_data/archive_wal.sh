#!/bin/bash

if [ "$#" -ne 2 ]; then
    exit 1
fi

PG_HOSTNAME=$(hostname)

gzip < /data/postgres/$1 > /data/postgres_archive/${PG_HOSTNAME}/$2.gz