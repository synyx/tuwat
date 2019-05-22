FROM php:7-apache

# moar ram
RUN { echo 'memory_limit=512M' > /usr/local/etc/php/conf.d/memory-limit.ini; }

COPY . /var/www/html/

