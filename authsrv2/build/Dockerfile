FROM ruby:3.0.1-alpine

WORKDIR /work

RUN apk --update add make postgresql-client build-base postgresql-dev tzdata bash

# install ruby dependencies
RUN gem install bundler:2.2.6
COPY authsrv2/Gemfile authsrv2/Gemfile.lock ./
RUN bundle install --jobs 20 --retry 5 --without "test development"

# copy app
COPY authsrv2/. ./

# setup cron jobs
RUN bundle exec whenever --update-crontab

COPY authsrv2/build/entrypoint.sh ./
ENTRYPOINT [ "./entrypoint.sh" ]
