FROM mdillon/postgis:11

COPY ./scripts/pginit.sh /docker-entrypoint-initdb.d/11_init.sh
