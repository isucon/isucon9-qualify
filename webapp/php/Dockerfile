#syntax=docker/dockerfile:1
FROM php:8.3-cli

RUN --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update -qq && apt-get install -y \
    git unzip default-mysql-client composer

WORKDIR /home/webapp

COPY init.sh /home/init.sh

COPY php/composer.json php/composer.lock ./
RUN composer install --no-dev

COPY php /home/webapp

ENTRYPOINT ["php", "-S", "0.0.0.0:8000", "-t", "public", "public/index.php"]
