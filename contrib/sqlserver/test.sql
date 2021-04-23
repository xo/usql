-- sqlserver test script

\set

\set SYNTAX_HL_FORMAT terminal16m
\set SYNTAX_HL true

\?

\copyright

\set SYNTAX_HL_STYLE dracula

select 'test''
' \g

\set NAME myname

DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS authors;

CREATE TABLE authors (
  author_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  name varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

CREATE TABLE books (
  book_id integer NOT NULL IDENTITY(1,1) PRIMARY KEY,
  author_id integer NOT NULL FOREIGN KEY REFERENCES authors(author_id),
  isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
  title varchar(255) NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available datetime2 NOT NULL DEFAULT CURRENT_TIMESTAMP,
  tags varchar(255) NOT NULL DEFAULT ''
);

CREATE INDEX books_title_idx ON books(title, year);

\set SYNTAX_HL_STYLE paraiso-dark

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

  /* exiting! */
\q
