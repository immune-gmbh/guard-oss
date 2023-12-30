#!/bin/bash

set -xe

# start PostgreSQL DB
mkdir -p /run/postgresql
chown -R postgres /run/postgresql
su - postgres -c "initdb /var/lib/postgresql/data"
echo "host all  all    0.0.0.0/0  md5" >> /var/lib/postgresql/data/pg_hba.conf
echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf
su - postgres -c "PGDATA=/var/lib/postgresql/data pg_ctl start"
psql -U postgres -c "CREATE DATABASE authsrv"

export SECRET_KEY_BASE=00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
export AUTHSRV_DATABASE=postgres
export AUTHSRV_DATABASE_USER=postgres
export AUTHSRV_STRIPE_SECRET_KEY=unset
export AUTHSRV_TWITTER_API_KEY=unset
export AUTHSRV_TWITTER_API_SECRET=unset
export AUTHSRV_GITHUB_API_KEY=unset
export AUTHSRV_GITHUB_API_SECRET=unset
export AUTHSRV_GITHUB_CALLBACK_URL=unset
export AUTHSRV_GOOGLE_API_KEY=unset
export AUTHSRV_GOOGLE_API_SECRET=unset
export AUTHSRV_GOOGLE_CALLBACK_URL=unset
export AUTHSRV_DUMP_SCHEMA_AS_RB=1
export AUTHSRV_JS_ROUTES_PATH="./authsrvRoutes.js"

# authsrvSchema.d.ts
bundler exec rails db:migrate
bundle exec schema2type -o ./authsrvSchema.d.ts
echo "export default schema;" >> ./authsrvSchema.d.ts

# authsrvRoutes.js
bundle exec rake js_routes_rails:export

# stop PostgreSQL
su - postgres -c "PGDATA=/var/lib/postgresql/data pg_ctl stop"
