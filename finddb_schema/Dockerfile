FROM mdillon/postgis
MAINTAINER Stefan Safranek, sjsafranek@gmail.com

RUN apt-get update
RUN apt-get install -y less htop vim telnet net-tools host

COPY database.sql /database.sql
COPY db_init.sh /docker-entrypoint-initdb.d/db_init.sh

RUN chown root:root /docker-entrypoint-initdb.d/db_init.sh && \
    chmod 4755 /docker-entrypoint-initdb.d/db_init.sh
