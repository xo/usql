\set POSTGRES_USER postgres
\set POSTGRES_PASS P4ssw0rd
\set POSTGRES_DB   postgres
\set POSTGRES_HOST `docker port postgres 5432 | head -n1`

\prompt NAME 'Create database user: '
\prompt -password PASS 'Password for "':NAME'": '

\connect 'postgres://':POSTGRES_USER':':POSTGRES_PASS'@':POSTGRES_HOST'/':POSTGRES_DB'?sslmode=disable'

DROP USER IF EXISTS :NAME;

CREATE USER :NAME PASSWORD :'PASS';

DROP DATABASE IF EXISTS :NAME;

CREATE DATABASE :NAME OWNER :NAME;
