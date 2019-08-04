#!/bin/bash

sudo -u postgres psql -c "CREATE USER findadmin WITH PASSWORD 'dev'"
sudo -u postgres psql -c "CREATE DATABASE find5"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE find5 to findadmin"
sudo -u postgres psql -c "ALTER USER findadmin WITH SUPERUSER"

PGPASSWORD=dev psql -d find5 -U findadmin -f database.sql
