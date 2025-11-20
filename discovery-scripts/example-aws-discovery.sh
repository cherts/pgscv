#!/usr/bin/env bash
#
# Discovery PostgreSQL instance from AWS RDS
#
# Usage: example-aws-discovery.sh --profile=[AWS_PROFILE] --region=[REGION]
#
# Example: example-aws-discovery.sh --profile=pgscv --region=us-west-1
#
# Author: Mikhail Grigorev

PSQL_USER="pgscv"
PSQL_PASSWORD="${PG_PASSWORD}"
PSQL_DB="postgres"
PSQL_ADDITIONAL_CONNSTR_OPTS="sslrootcert=aws-global-bundle.pem&sslmode=verify-full"

## Function to display usage information
usage() {
  cat << EOF
Usage: $0 [OPTIONS]

A demo script discovery PostgreSQL instance from AWS RDS

Options:
  -p, --profile PROFILE  AWS CLI profile name (required)
  -r, --region REGION    AWS region name (default: us-west-1)
  -v, --verbose          Enable verbose output
  -h, --help             Display this help message

Examples:
  $0 -p default -r us-west-1
  $0 --profile=default --region=us-west-1 
  $0 --profile=default --region=us-west-1 --verbose
EOF
  exit 1
}

## Print verbose message
verbose() {
	MSG=$1
	if [ "${VERBOSE}" = true ]; then
		echo "${MSG}"
	fi
}

## Parse command-line options
OPTS=$(getopt -o p:r:vh --long profile:,region:,verbose,help -n 'example-aws-discovery.sh' -- "$@")

if [ $? -ne 0 ]; then
  echo "Failed to parse options" >&2
  usage
fi

## Reset the positional parameters to the parsed options
eval set -- "$OPTS"

## Initialize variables with default values
AWS_CLI_PROFILE=""
AWS_REGION="eu-north-1"
VERBOSE=false

## Process the options
while true; do
  case "$1" in
    -p | --profile)
      AWS_CLI_PROFILE="$2"
      shift 2
      ;;
    -r | --region)
      AWS_REGION="$2"
      shift 2
      ;;
    -v | --verbose)
      VERBOSE=true
      shift
      ;;
    -h | --help)
      usage
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "Internal error!"
      exit 1
      ;;
  esac
done

## Check if required options are provided
if [ -z "${AWS_CLI_PROFILE}" ]; then
  echo "Error: AWS profile name must be specified with -p or --profile option." >&2
  usage
fi

# Check command exist function
_command_exists() {
	type "$1" &>/dev/null
}

if _command_exists aws; then
	AWS_BIN=$(which aws)
else
	echo "Error: aws cli not found, see https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-install.html" >&2
	exit 2
fi

verbose "Using AWS CLI binary: ${AWS_BIN}"
verbose "AWS CLI profile: ${AWS_CLI_PROFILE}"
verbose "AWS region: ${AWS_REGION}"

${AWS_BIN} configure get region --profile ${AWS_CLI_PROFILE} >/dev/null 2>&1
if [[ $? -ne 0 ]]; then
	verbose "Error: AWS CLI profile '${AWS_CLI_PROFILE}' not found or not configured." >&2
	exit 2
fi

#${AWS_BIN} sts get-caller-identity >/dev/null 2>&1
#if [[ $? -ne 0 ]]; then
#	verbose "Error: AWS CLI profile '${AWS_CLI_PROFILE}' has invalid credentials." >&2
#	exit 2
#fi

discover_databases() {
	local HOST=$1
	local PORT=${2:-"5432"}
	local SERVICE_ID=$3

	local DB_LIST=$(psql -X -t -A "postgres://${PSQL_USER}:${PSQL_PASSWORD}@${HOST}:${PORT}/${PSQL_DB}?${PSQL_ADDITIONAL_CONNSTR_OPTS}" -c "
        SELECT datname
        FROM pg_database
        WHERE datistemplate = false
        AND datname NOT IN ('postgres', 'template0', 'template1', 'rdsadmin')
    " postgres 2>/dev/null || true)

	if [ -z "${DB_LIST}}" ]; then
		return
	fi

	while IFS= read -r DB; do
		[ -z "${DB}" ] && continue
		echo "${SERVICE_ID}-${DB} postgres://${PSQL_USER}@${HOST}:${PORT}/${DB}?${PSQL_ADDITIONAL_CONNSTR_OPTS}"
	done <<<"${DB_LIST}"
}

echo "# service-id dsn password-from-env"

AWS_INSTANCE_LIST=$(${AWS_BIN} rds describe-db-instances \
	--profile "${AWS_CLI_PROFILE}" \
	--region "${AWS_REGION}" \
	--query "DBInstances[?Engine=='postgres'|'aurora-postgresql'].[DBInstanceIdentifier,DBInstanceStatus,Endpoint.Address,Endpoint.Port]" \
	--output text 2>/dev/null)

IFS=$'\n'
for PG in ${AWS_INSTANCE_LIST[@]}; do
	INSTANCE=$(echo ${PG} | awk '{print $1}')
	STATUS=$(echo ${PG} | awk '{print $2}')
	HOST=$(echo ${PG} | awk '{print $3}')
	PORT=$(echo ${PG} | awk '{print $4}')
	if [[ "${STATUS}" != "available" ]]; then
		continue
	fi
	if [ -z "${PORT}" ]; then
		PORT=5432
	fi
	if [ -z "${HOST}" ]; then
		continue
	fi
	#echo "svc-pg-${PORT} postgres://${PSQL_USER}@${HOST}:${PORT}/${PSQL_DB}?${PSQL_ADDITIONAL_CONNSTR_OPTS} PG_PASSWORD"
	discover_databases "${HOST}" "${PORT}" "svc-pg-${INSTANCE}" | while IFS= read -r SVC; do
		echo "${SVC} PG_PASSWORD"
	done
done