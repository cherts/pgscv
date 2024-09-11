CREATE USER pgscv WITH NOCREATEDB NOCREATEROLE LOGIN PASSWORD 'pgscv';
GRANT pg_read_server_files, pg_monitor TO pgscv;
GRANT EXECUTE on FUNCTION pg_current_logfile() TO pgscv;
CREATE USER pgbench WITH NOCREATEDB NOCREATEROLE LOGIN PASSWORD 'pgbench';
CREATE DATABASE pgbench WITH OWNER = pgbench;
GRANT ALL PRIVILEGES ON DATABASE pgbench TO pgbench;
\c pgbench
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
