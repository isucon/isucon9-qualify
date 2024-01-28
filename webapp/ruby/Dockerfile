FROM ruby:3.3-slim

RUN apt-get update && apt-get install -y \
  build-essential \
  default-mysql-client \
  default-libmysqlclient-dev \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /home/webapp

COPY init.sh /home/init.sh

COPY ruby/Gemfile /home/webapp
COPY ruby/Gemfile.lock /home/webapp
WORKDIR /home/webapp
RUN bundle config set --local path 'vendor/bundle'
RUN bundle install
COPY ruby /home/webapp

CMD bundle exec rackup -p 8000 -o 0.0.0.0
