-- sqlite3 test script

\set

help

\?

\copyright

\set SYNTAX_HL_STYLE dracula

select 'test''
' \g

\set NAME myname

PRAGMA foreign_keys = 1;

DROP TABLE IF EXISTS books;

DROP TABLE IF EXISTS authors;

CREATE TABLE authors (
  author_id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  name text NOT NULL DEFAULT ''
);

CREATE INDEX authors_name_idx ON authors(name);

\set SYNTAX_HL_STYLE paraiso-dark

CREATE TABLE books (
  /*
    this is a multiline comment
   */
  book_id integer NOT NULL PRIMARY KEY AUTOINCREMENT, -- the id of the author
  author_id integer NOT NULL REFERENCES authors(author_id),
  isbn text NOT NULL DEFAULT '' UNIQUE,
  title text NOT NULL DEFAULT '',
  year integer NOT NULL DEFAULT 2000,
  available timestamp with time zone NOT NULL DEFAULT '',
  tags text NOT NULL DEFAULT '{}'
);

CREATE INDEX books_title_idx ON books(title, year);

insert into authors (name) values
  ("jk rowling"),
  ("author amazing")
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
