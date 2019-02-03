\set POSTGRES_USER postgres
\set POSTGRES_PASS P4ssw0rd
\set POSTGRES_DB   postgres
\set POSTGRES_HOST `docker port postgres 5432`

\prompt 'Database user: ' NAME
\prompt 'Database pass: ' PASS

\connect 'postgres://':POSTGRES_USER':':POSTGRES_PASS'@':POSTGRES_HOST'/':POSTGRES_DB

DROP USER IF EXISTS :NAME;

CREATE USER :NAME PASSWORD :'PASS';

DROP DATABASE IF EXISTS :NAME;

CREATE DATABASE :NAME OWNER :NAME;
