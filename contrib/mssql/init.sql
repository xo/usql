\set MSSQL_USER sa
\set MSSQL_PASS Adm1nP@ssw0rd
\set MSSQL_DB   postgres
\set MSSQL_HOST `docker port mssql 1433`

\prompt NAME 'Create database user: '
\prompt -password PASS 'Password for "':NAME'": '

\connect 'mssql://':MSSQL_USER':':MSSQL_PASS'@':MSSQL_HOST'/':MSSQL_DB

EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN :NAME;

DROP DATABASE :NAME;

CREATE DATABASE :NAME
  containment=partial;

\connect 'mssql://':MSSQL_USER':':MSSQL_PASS'@':MSSQL_HOST'/':MSSQL_DB

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
