#!/bin/sh

set -eu -o pipefail

RUN_MIGRATIONS=0
START_RAILS=0
START_CRON=0

echo "args: $@"

for I in "$@"
do
  case "$I" in

  --run-migrations)
    RUN_MIGRATIONS=1
    ;;

  --start-rails)
    START_RAILS=1
    ;;

  --start-cron)
    START_CRON=1
    ;;

  *)
    echo "Unknown arg "$I". Ignoring"
    ;;

  esac
done

if [[ "$RUN_MIGRATIONS" == "0" && "$START_RAILS" == "0" && "$START_CRON" == "0" ]]
then
  echo "No args given. Default to --start"
  START_RAILS=1
fi

if [[ "$START_RAILS" == "1" && "$START_CRON" == "1" ]]
then
  echo "Cannot run cron and rails."
  exit 1
fi

if [[ "$RUN_MIGRATIONS" == "1" ]]
then
  # wait for database to come online
  bundler exec rails db:wait

  # run migrations
  bundler exec rails db:migrate

  # fix permissions
  bundler exec rails "db:chown[$AUTHSRV_DATABASE_USER_ROLE]"

  # seed database
  #bundler exec rails db:seed
fi

if [[ "$START_RAILS" == "1" ]]
then
  # start server
  exec bundler exec rails server --log-to-stdout
fi

if [[ "$START_CRON" == "1" ]]
then
  # start crond
  exec crond -f -l 8
fi
