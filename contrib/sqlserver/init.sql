\set SQLSERVER_USER sa
\set SQLSERVER_PASS Adm1nP@ssw0rd
\set SQLSERVER_DB   postgres
\set SQLSERVER_HOST `docker port sqlserver 1433`

\prompt NAME 'Create database user: '
\prompt -password PASS 'Password for "':NAME'": '

\connect 'sqlserver://':SQLSERVER_USER':':SQLSERVER_PASS'@':SQLSERVER_HOST'/':SQLSERVER_DB

EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN :NAME;

DROP DATABASE :NAME;

CREATE DATABASE :NAME
  containment=partial;

\connect 'sqlserver://':SQLSERVER_USER':':SQLSERVER_PASS'@':SQLSERVER_HOST'/':SQLSERVER_DB

CREATE LOGIN :NAME
  WITH
    password=:'PASS',
    check_policy=off,
    default_database=:NAME;

CREATE USER :NAME
  FOR login :NAME
  WITH default_schema=:NAME;

CREATE SCHEMA :NAME authorization :NAME;

EXEC sp_addrolemember
  'db_owner', :'NAME';
