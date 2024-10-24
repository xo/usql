EXEC sp_configure
  'contained database authentication', 1;

RECONFIGURE;

DROP LOGIN :NAME;

DROP DATABASE :NAME;

CREATE DATABASE :NAME
  CONTAINMENT=PARTIAL;

\set QNAME "''":NAME"''"

\set SQL 'CREATE LOGIN ':NAME' WITH PASSWORD=':QNAME', CHECK_POLICY=OFF, DEFAULT_DATABASE=':NAME';'
EXEC [:NAME].[dbo].[sp_executesql] N:'SQL'

\set SQL 'CREATE USER ':NAME' FOR LOGIN ':NAME' WITH DEFAULT_SCHEMA=':NAME';'
EXEC [:NAME].[dbo].[sp_executesql] N:'SQL';

\set SQL 'CREATE SCHEMA ':NAME' AUTHORIZATION ':NAME';'
EXEC [:NAME].[dbo].[sp_executesql] N:'SQL';

\set SQL 'EXEC sp_addrolemember db_owner, ':QNAME';'
EXEC [:NAME].[dbo].[sp_executesql] N:'SQL';

-- original reconnect version:
--
--\connect 'sqlserver://localhost/':NAME
--
--CREATE LOGIN :NAME
--  WITH
--    PASSWORD=:'PASS',
--    CHECK_POLICY=OFF,
--    DEFAULT_DATABASE=:NAME;
--
--CREATE USER :NAME
--  FOR LOGIN :NAME
--  WITH DEFAULT_SCHEMA=:NAME;
--
--CREATE SCHEMA :NAME AUTHORIZATION :NAME;
--
--EXEC sp_addrolemember 'db_owner', :'NAME';
