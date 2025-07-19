#syntax=docker/dockerfile:1

FROM php:8.4-fpm-alpine

RUN apk add --no-cache \
    $PHPIZE_DEPS \
    && docker-php-ext-install pdo pdo_mysql \
    && pecl install apcu \
    && docker-php-ext-enable apcu

RUN mkdir -p /home/webapp

COPY init.sh /home/init.sh

WORKDIR /home/webapp

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY php/composer.json ./
RUN composer install --no-dev --optimize-autoloader

COPY php /home/webapp

RUN chown -R www-data:www-data /home/webapp

EXPOSE 9000

CMD ["php-fpm"]