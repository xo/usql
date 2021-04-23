-- mysql test script

\set

\set SYNTAX_HL_FORMAT terminal16m
\set SYNTAX_HL true

\?

\copyright

\set SYNTAX_HL_STYLE dracula

select 'test''
' \g

\set NAME myname

drop database if exists testdb; create database testdb; use testdb;

SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS books;
DROP FUNCTION IF EXISTS say_hello;
SET FOREIGN_KEY_CHECKS=1;

CREATE TABLE authors (
  author_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name text NOT NULL DEFAULT ''
) ENGINE=InnoDB;

CREATE INDEX authors_name_idx ON authors(name(255));

\set SYNTAX_HL_STYLE paraiso-dark

CREATE TABLE books (
  /*
    this is a multiline comment
   */
  book_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
  author_id integer NOT NULL,
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  book_type ENUM('FICTION', 'NONFICTION') NOT NULL DEFAULT 'FICTION',
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available datetime NOT NULL DEFAULT NOW(),
  tags text NOT NULL DEFAULT '',
  CONSTRAINT FOREIGN KEY (author_id) REFERENCES authors(author_id)
) ENGINE=InnoDB;

CREATE INDEX books_title_idx ON books(title, year);

insert into authors (name) values
  ('jk rowling'),
  ('author amazing')
\g

  select * from authors;

\set COLNAME name
\set NAME amaz

\echo `echo hello`

select :"COLNAME" from authors where :COLNAME like '%' || :'NAME' || '%'

\print \raw

\g

\gset AUTHOR_

select :'AUTHOR_name';

\begin
insert into authors (name) values ('test');
\rollback

insert into authors (name) values ('hello');
select * from authors;

insert into books (author_id, isbn, title, year, available) values
  (1, '1', 'one', 2018, '2018-06-01 00:00:00'),
  (2, '2', 'two', 2019, '2019-06-01 00:00:00')
;

select * from books b inner join authors a on a.author_id = b.author_id;

CREATE FUNCTION say_hello(s text) RETURNS text
  DETERMINISTIC
  RETURN CONCAT('hello ', s);

select say_hello('a name!') \G

  /* exiting! */
\q
