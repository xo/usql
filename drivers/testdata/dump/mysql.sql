-- The fixture of TestDump for MySQL and MariaDB. It holds one of each object
-- and type that a dump writes, and rows that test the literals.
CREATE TABLE authors (
  author_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(200) NOT NULL,
  email VARCHAR(100) UNIQUE,
  born DATE,
  rating DECIMAL(4,2) DEFAULT 2.50,
  active TINYINT(1) NOT NULL DEFAULT 1,
  mood ENUM('sad','ok','happy'),
  photo VARBINARY(64),
  doc BLOB,
  notes TEXT,
  profile JSON,
  created DATETIME(6) NOT NULL DEFAULT '2000-01-01 00:00:00',
  score DOUBLE,
  small SMALLINT,
  big BIGINT UNSIGNED,
  initials CHAR(2),
  flags SET('a','b','c'),
  name_upper VARCHAR(200) GENERATED ALWAYS AS (UPPER(name)) STORED,
  CONSTRAINT authors_rating_check CHECK (rating >= 0)
) COMMENT 'People who write books';
CREATE TABLE books (
  book_id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  author_id INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  isbn VARCHAR(20),
  price DECIMAL(10,2),
  published DATETIME,
  UNIQUE KEY books_isbn_key (isbn),
  KEY books_author_published_idx (author_id, published),
  CONSTRAINT books_author_id_fkey FOREIGN KEY (author_id) REFERENCES authors (author_id) ON DELETE CASCADE
);
CREATE INDEX books_title_idx ON books (title);
CREATE TABLE nokey (a TEXT, b INT);
-- A TIMESTAMP is an instant, which the client reads in the time zone of the
-- session. It has a table of its own, so that a known difference of the
-- comparison with mysqldump hides one row only.
CREATE TABLE stamps (stamp_id INT NOT NULL PRIMARY KEY, stamp TIMESTAMP NULL);
CREATE VIEW book_list AS SELECT b.title, a.name FROM books b JOIN authors a ON a.author_id = b.author_id;
INSERT INTO authors (name, email, born, rating, active, mood, photo, doc, notes, profile, created, score, small, big, initials, flags) VALUES
  ('Ann ''Quote''', 'ann@example.com', '1970-01-02', 4.25, 1, 'happy', X'00ff10', X'deadbeef', 'Tab	here
newline \\ backslash', '{"k": [1, 2], "s": "x\\"y"}', '2024-02-29 12:34:56.789012', 1.5e300, -32768, 18446744073709551615, 'AB', 'a,c'),
  ('Bøb ünicode 🙂', NULL, NULL, NULL, 0, NULL, NULL, NULL, NULL, NULL, '1999-12-31 23:59:59', NULL, NULL, NULL, NULL, ''),
  ('Zed', NULL, '1000-01-01', 0, 1, 'sad', X'', X'', '', '[]', '2000-01-01 00:00:00', 0, 0, 0, 'Z', NULL);
INSERT INTO books (author_id, title, isbn, price, published) VALUES
  (1, 'First; with semicolon', '978-0', 12.50, '2020-05-06 07:08:09'),
  (2, 'Line
break', NULL, NULL, NULL),
  (3, '$$ dollar $$ quoted', '978-1', 0.01, '1999-01-01 00:00:00');
INSERT INTO nokey VALUES ('x', 1), (NULL, NULL), ('x', 1);
INSERT INTO stamps VALUES (1, '2024-02-29 12:34:56'), (2, NULL);
