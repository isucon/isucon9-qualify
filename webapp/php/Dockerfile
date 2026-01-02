#syntax=docker/dockerfile:1

FROM php:8.5-fpm-trixie

RUN apt-get update && apt-get install -y \
    libzip-dev \
    zip \
    unzip \
    default-mysql-client \
    && docker-php-ext-install pdo pdo_mysql \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /home/webapp && \
    chown -R www-data:www-data /home/webapp && \
    chmod -R 755 /home/webapp

COPY init.sh /home/init.sh
RUN chmod +x /home/init.sh

WORKDIR /home/webapp

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY php/composer.json ./
RUN COMPOSER_CACHE_DIR=/tmp/composer-cache composer install --no-dev --optimize-autoloader

COPY php /home/webapp

EXPOSE 9000

CMD ["php-fpm"]
