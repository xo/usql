-- The fixture of TestDump for SQLite. It holds one of each object and type
-- that a dump writes, and rows that test the literals.
CREATE TABLE authors (
  author_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  born DATE,
  rating REAL DEFAULT 1.5,
  active BOOLEAN NOT NULL DEFAULT 1,
  photo BLOB,
  price NUMERIC(10,2),
  note VARCHAR(20)
);
CREATE TABLE books (
  book_id INTEGER PRIMARY KEY,
  author_id INTEGER NOT NULL REFERENCES authors (author_id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  published TEXT
);
CREATE INDEX books_title_idx ON books (title);
-- There is no unique index, because dbmeta writes one as an unnamed UNIQUE
-- inside CREATE TABLE, and the copy then lists it in another place. Reported
-- to dbmeta on 2026-11-08.
CREATE TABLE nokey (a TEXT, b INTEGER);
CREATE VIEW book_list AS SELECT b.title, a.name FROM books b JOIN authors a ON a.author_id = b.author_id;
CREATE TRIGGER books_ai AFTER INSERT ON books BEGIN UPDATE authors SET note = 'has books' WHERE author_id = NEW.author_id; END;
INSERT INTO authors (author_id, name, born, rating, active, photo, price, note) VALUES
  (1, 'Ann ''Quote''', '1970-01-02', 4.25, 1, X'00ff10', 12.5, NULL),
  (2, 'Bøb ünicode 🙂', NULL, NULL, 0, NULL, NULL, NULL),
  (3, 'Tab	here
newline \ backslash', '0001-01-01', 0, 1, X'', 0.01, '');
INSERT INTO books VALUES (1, 1, 'First; with semicolon', '2020-05-06 07:08:09'), (2, 2, 'Line
break', NULL), (3, 3, '$$ dollar $$ quoted', '1999-01-01');
INSERT INTO nokey VALUES ('x', 1), (NULL, NULL), ('x', 1);
