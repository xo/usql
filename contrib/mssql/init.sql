\connect mssql://sa:Adm1nP@ssw0rd@localhost/

\prompt `Database user: ` NAME

EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN :NAME;

DROP DATABASE :NAME;

CREATE DATABASE :NAME
  containment=partial;

\connect 'mssql://sa:Adm1nP@ssw0rd@localhost/':NAME

CREATE LOGIN :NAME
  WITH
    password=:'NAME',
    check_policy=off,
    default_database=:NAME;

CREATE USER :NAME
  FOR login :NAME
  WITH default_schema=:NAME;

CREATE SCHEMA :NAME authorization :NAME;

EXEC sp_addrolemember
  'db_owner', :'NAME';
