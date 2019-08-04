#!/bin/bash

sudo -u postgres psql -c "CREATE USER finduser WITH PASSWORD 'dev'"
sudo -u postgres psql -c "CREATE DATABASE finddb"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE finddb to finduser"
sudo -u postgres psql -c "ALTER USER finduser WITH SUPERUSER"

PGPASSWORD=dev psql -d finddb -U finduser -f database.sql
