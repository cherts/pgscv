CREATE USER pgscv WITH NOCREATEDB NOCREATEROLE LOGIN PASSWORD 'pgscv';
GRANT pg_read_server_files, pg_monitor TO pgscv;
GRANT EXECUTE on FUNCTION pg_current_logfile() TO pgscv;
CREATE USER pgbench WITH NOCREATEDB NOCREATEROLE LOGIN PASSWORD 'pgbench';
SELECT 'CREATE DATABASE pgbench WITH OWNER = pgbench' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'pgbench')\gexec
GRANT ALL PRIVILEGES ON DATABASE pgbench TO pgbench;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE USER repluser WITH NOCREATEDB NOCREATEROLE LOGIN REPLICATION PASSWORD 'repluser';
SELECT pg_create_physical_replication_slot('replica_slot1');
\c pgbench
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
