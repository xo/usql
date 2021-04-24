\set PGDB pg://postgres:P4ssw0rd@localhost
\set MYDB my://root:P4ssw0rd@localhost
\set SQDB sq://test3.db
\set MSDB ms://sa:Adm1nP@ssw0rd@localhost/

\connect :PGDB
drop table if exists a_bit_of_everything;
create table a_bit_of_everything (
  a_id serial primary key,
  a_blob bytea,
  a_bool boolean,
  a_date timestamp with time zone,
  a_double double precision,
  a_int integer,
  a_text text
);
insert into a_bit_of_everything
  (a_blob, a_bool, a_date, a_double, a_int, a_text)
values
  (E'more\ntext'::bytea, true, now(), 32.0, 0, 'some text'),
  (E'other\ntext'::bytea, false, now()+interval '3 days', 64.0, 128, 'foobar')
;
select * from a_bit_of_everything;

\connect :MYDB
drop database if exists testdb;
create database testdb;
use testdb;
drop table if exists a_bit_of_everything;
create table a_bit_of_everything (
  a_id integer not null auto_increment primary key,
  a_blob blob,
  a_bool boolean,
  a_date datetime,
  a_double double,
  a_int integer,
  a_text text
);
\copy :PGDB :MYDB/testdb 'select * from a_bit_of_everything' 'a_bit_of_everything(a_id, a_blob, a_bool, a_date, a_double, a_int, a_text)'
\connect :MYDB/testdb
select * from a_bit_of_everything;

\! rm -f test3.db
\connect :SQDB
create table a_bit_of_everything (
  a_id integer primary key autoincrement,
  a_blob blob,
  a_bool boolean,
  a_date datetime,
  a_double double precision,
  a_int integer,
  a_text text
);
\copy :PGDB :SQDB 'select * from a_bit_of_everything' 'a_bit_of_everything(a_id, a_blob, a_bool, a_date, a_double, a_int, a_text)'
\connect :SQDB
select * from a_bit_of_everything;

\connect :MSDB
drop table  if exists a_bit_of_everything;
create table a_bit_of_everything (
  -- a_id integer identity(1,1) primary key, -- doesn't work currently
  a_id integer,
  a_blob varbinary(max),
  a_bool bit,
  a_date datetime2,
  a_double double precision,
  a_int integer,
  a_text text
);
\copy :PGDB :MSDB 'select * from a_bit_of_everything' 'a_bit_of_everything(a_id, a_blob, a_bool, a_date, a_double, a_int, a_text)'
\connect :MSDB
select * from a_bit_of_everything;

\quit
