-- The fixture of TestDump for DuckDB. It holds one of each object and type
-- that a dump writes, and rows that test the literals.
CREATE SCHEMA app;
CREATE SEQUENCE app.author_seq START 1;
CREATE TABLE app.authors (
  author_id INTEGER PRIMARY KEY DEFAULT nextval('app.author_seq'),
  name VARCHAR NOT NULL UNIQUE,
  born DATE,
  rating DECIMAL(4,2) DEFAULT 2.50 CHECK (rating >= 0),
  active BOOLEAN NOT NULL DEFAULT true,
  photo BLOB,
  ident UUID,
  created TIMESTAMP WITH TIME ZONE,
  local TIMESTAMP,
  score DOUBLE,
  big HUGEINT,
  small SMALLINT
);
CREATE TABLE app.books (
  book_id INTEGER PRIMARY KEY,
  author_id INTEGER NOT NULL REFERENCES app.authors (author_id),
  title VARCHAR NOT NULL,
  price DECIMAL(10,2)
);
CREATE INDEX books_title_idx ON app.books (title);
CREATE TABLE main.nokey (a VARCHAR, b INTEGER);
CREATE VIEW app.book_list AS SELECT b.title, a.name FROM app.books b JOIN app.authors a ON a.author_id = b.author_id;
INSERT INTO app.authors (name, born, rating, active, photo, ident, created, local, score, big, small) VALUES
  ('Ann ''Quote''', '1970-01-02', 4.25, true, '\x00\xFF\x10'::BLOB, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2024-02-29 12:34:56.789012+00', '2024-02-29 12:34:56.789012', 1.5e300, 170141183460469231731687303715884105727, -32768),
  ('Bøb ünicode 🙂', NULL, NULL, false, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
  ('Tab	here
newline \ backslash', '0001-01-01', 0, true, ''::BLOB, NULL, '2000-01-01 00:00:00+00', '2000-01-01 00:00:00', 0, -1, 0);
INSERT INTO app.books VALUES (1, 1, 'First; with semicolon', 12.50), (2, 2, 'Line
break', NULL), (3, 3, '$$ dollar $$ quoted', 0.01);
INSERT INTO main.nokey VALUES ('x', 1), (NULL, NULL), ('x', 1);
