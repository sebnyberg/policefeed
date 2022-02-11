FROM mdillon/postgis:12

COPY ./scripts/pginit.sh /docker-entrypoint-initdb.d/11_init.sh
