#!/bin/bash

psql -c "CREATE USER finduser WITH PASSWORD 'dev'"
psql -c "CREATE DATABASE finddb"
psql -c "GRANT ALL PRIVILEGES ON DATABASE finddb to finduser"
psql -c "ALTER USER finduser WITH SUPERUSER"

# PGPASSWORD=dev psql -d finddb -U finduser -f database.sql

cd base_schema
PGPASSWORD=dev psql -d finddb -U finduser -f db_setup.sql
cd ..
