\connect odbc+db2://db2inst1:P4ssw0rd@localhost/testdb

create schema test;

create table test.mytable (
  COL1 INTEGER NOT NULL,
  COL2 CHAR(25),
  COL3 VARCHAR(25) NOT NULL,
  COL4 DATE,
  COL5 DECIMAL(10,2),
  PRIMARY KEY (COL1),
  UNIQUE (COL3)
);

insert into test.mytable
  (col1, col2, col3, col4, col5)
values
  (1, 'a', 'first', current date, 15.0),
  (2, 'b', 'second', current date, 16.0),
  (3, 'c', 'third', current date, 17.0)
;

select * from test.mytable;
