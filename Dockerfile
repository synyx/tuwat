FROM registry.synyx.cloud/docker.io/library/php:7-apache

RUN apt-get update \
   && apt-get install -y perl bash ddate libjson-perl libwww-perl libsys-syslog-perl liblockfile-simple-perl


# moar ram
RUN { echo 'memory_limit=512M' > /usr/local/etc/php/conf.d/memory-limit.ini; }

# Setup localization
ENV LANG="de_DE.UTF-8" TZ="Europe/Berlin"

#RUN yes | pecl install xdebug \
#    && echo "zend_extension=$(find /usr/local/lib/php/extensions/ -name xdebug.so)" > /usr/local/etc/php/conf.d/xdebug.ini \
#    && echo "xdebug.idekey = PHPSTORM" >> /usr/local/etc/php/conf.d/xdebug.ini \
#    && echo "xdebug.remote_enable=on" >> /usr/local/etc/php/conf.d/xdebug.ini \
#    && echo "xdebug.remote_autostart=off" >> /usr/local/etc/php/conf.d/xdebug.ini

COPY . /var/www/html/
