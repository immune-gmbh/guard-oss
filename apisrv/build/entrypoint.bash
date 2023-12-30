#!/bin/bash
set -eu -o pipefail

RUN_MIGRATIONS=0
START_SERVICE=0

echo "args: $@"

for I in "$@"
do
  case "$I" in

  --run-migrations)
    RUN_MIGRATIONS=1
    ;;

  --start)
    START_SERVICE=1
    ;;

  *)
    echo "Unknown arg "$I". Ignoring"
    ;;

  esac
done

if [[ "$RUN_MIGRATIONS" == "0" && "$START_SERVICE" == "0" ]]
then
  echo "No args given. Default to --run-migrations and --start"
  RUN_MIGRATIONS=1
  START_SERVICE=1
fi

if [[ "$RUN_MIGRATIONS" == "1" ]]
then
  ADMIN_PASSWORD="$APISRV_DATABASE_ADMIN_PWD" USER_PASSWORD="$APISRV_DATABASE_USER_PWD" /migration \
    -database-url "$APISRV_DATABASE_URL" \
    -database-name "$APISRV_DATABASE_NAME" \
    -database-cert "$APISRV_DATABASE_SSLCERT" \
    -wait
fi

if [[ "$START_SERVICE" == "1" ]]
then
  exec /apisrv
fi
