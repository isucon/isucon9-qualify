#syntax=docker/dockerfile:1

FROM ruby:3.3-slim

RUN \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  --mount=type=cache,target=/var/cache/apt,sharing=locked \
  apt-get update -qq && apt-get install -y \
  build-essential \
  default-mysql-client \
  default-libmysqlclient-dev

RUN mkdir -p /home/webapp

COPY init.sh /home/init.sh

WORKDIR /home/webapp

COPY ruby/Gemfile ./
COPY ruby/Gemfile.lock ./
RUN bundle config set --local path 'vendor/bundle'
RUN bundle install

COPY ruby /home/webapp

ENTRYPOINT [ "bundle", "exec", "rackup", "-p", "8000", "-o", "0.0.0.0" ]
