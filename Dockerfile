FROM php:7-apache

RUN apt-get update \
   && apt-get install -y perl bash ddate libjson-perl libwww-perl libsys-syslog-perl liblockfile-simple-perl


# moar ram
RUN { echo 'memory_limit=512M' > /usr/local/etc/php/conf.d/memory-limit.ini; }

COPY . /var/www/html/
