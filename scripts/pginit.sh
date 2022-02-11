#!/usr/bin/env bash

set -eux

export PGUSER="$POSTGRES_USER"

database="dev"
schema="dev"

# Create the database
psql --dbname="postgres" <<EOM
DROP DATABASE IF EXISTS "${database}";
CREATE DATABASE "${database}";
EOM

psql --dbname="${database}" <<EOM
CREATE SCHEMA ${schema};
ALTER DATABASE dev SET search_path TO "\$user",${schema},public;
EOM
