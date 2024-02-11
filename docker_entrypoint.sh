#!/bin/sh

ARG1=$1

case "${ARG1}" in
"bash" | "sh")
  echo ${ARG1}
  exec "$@"
  ;;
*)
  exec /bin/pgscv "$@"
  ;;
esac
