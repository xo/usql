-- The fixture of TestDump for PostgreSQL. It holds one of each object and
-- type that a dump writes, and rows that test the literals.
CREATE SCHEMA app;
CREATE TYPE app.mood AS ENUM ('sad', 'ok', 'happy');
CREATE SEQUENCE app.ticket_seq START 100 INCREMENT 5;
CREATE TABLE app.authors (
  author_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name text NOT NULL,
  email character varying(100) UNIQUE,
  born date,
  rating numeric(4,2) DEFAULT 2.50 CHECK (rating >= 0),
  active boolean NOT NULL DEFAULT true,
  mood app.mood,
  photo bytea,
  tags text[],
  profile jsonb,
  ident uuid,
  created timestamp with time zone NOT NULL DEFAULT now(),
  wake time without time zone,
  span interval,
  score double precision,
  small smallint,
  big bigint,
  initials character(2),
  name_upper text GENERATED ALWAYS AS (upper(name)) STORED
);
COMMENT ON TABLE app.authors IS 'People who write books';
COMMENT ON COLUMN app.authors.name IS 'The full name';
CREATE TABLE app.books (
  book_id serial PRIMARY KEY,
  author_id integer NOT NULL REFERENCES app.authors (author_id) ON DELETE CASCADE,
  title text NOT NULL,
  isbn character varying(20),
  price numeric(10,2),
  published timestamp without time zone,
  ticket integer DEFAULT nextval('app.ticket_seq'),
  CONSTRAINT books_isbn_key UNIQUE (isbn)
);
CREATE INDEX books_title_idx ON app.books (lower(title));
CREATE INDEX books_author_published_idx ON app.books (author_id, published DESC) WHERE published IS NOT NULL;
CREATE TABLE public.nokey (a text, b integer);
CREATE VIEW app.book_list AS
  SELECT b.title, a.name FROM app.books b JOIN app.authors a ON a.author_id = b.author_id;
CREATE VIEW app.recent AS
  SELECT title FROM app.book_list WHERE title <> '';
CREATE FUNCTION app.touch() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  NEW.title := trim(NEW.title);
  RETURN NEW;
END;
$$;
CREATE TRIGGER books_touch BEFORE INSERT OR UPDATE ON app.books FOR EACH ROW EXECUTE FUNCTION app.touch();
INSERT INTO app.authors (name, email, born, rating, active, mood, photo, tags, profile, ident, created, wake, span, score, small, big, initials) VALUES
  ('Ann ''Quote''', 'ann@example.com', '1970-01-02', 4.25, true, 'happy', '\x00ff10', ARRAY['a','b c','d"e'], '{"k": [1, 2], "s": "x\"y"}', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2024-02-29 12:34:56.789012+05:30', '07:30:00', '1 year 2 mons 3 days 04:05:06', 1.5e300, -32768, 9223372036854775807, 'AB'),
  ('Bøb ünicode 🙂', NULL, NULL, NULL, false, NULL, NULL, '{}', 'null', NULL, '2000-01-01 00:00:00+00', NULL, NULL, 2.5, NULL, NULL, NULL),
  (E'Tab\there\nnewline \\ backslash', NULL, '0001-01-01', 0, true, 'sad', '\x', NULL, '[]', NULL, '1999-12-31 23:59:59+00', '23:59:59.999999', '-1 day', -0.5, 0, -1, 'Z');
INSERT INTO app.books (author_id, title, isbn, price, published) VALUES
  (1, '  First; with semicolon  ', '978-0', 12.50, '2020-05-06 07:08:09.123456'),
  (2, 'Line
break', NULL, NULL, NULL),
  (3, '$$ dollar $$ quoted', '978-1', 0.01, '1999-01-01 00:00:00');
SELECT nextval('app.ticket_seq');
-- The values of PostgreSQL that some targets cannot hold. A dump to SQLite
-- leaves the table out.
CREATE TABLE app.floats (float_id integer PRIMARY KEY, f double precision, r real);
INSERT INTO app.floats VALUES (1, 'NaN', 'NaN'), (2, 'Infinity', 'Infinity'), (3, '-Infinity', '-Infinity'), (4, -0.0, 1.5);
INSERT INTO public.nokey VALUES ('x', 1), (NULL, NULL), ('x', 1);
-- A partitioned table, with rows in each partition, and a materialized view
-- that reads it. dbmeta dumps both from v0.40.0 (W77).
CREATE TABLE app.events (
  event_id integer NOT NULL,
  happened date NOT NULL,
  kind text,
  PRIMARY KEY (event_id, happened)
) PARTITION BY RANGE (happened);
CREATE TABLE app.events_2025 PARTITION OF app.events FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE app.events_2026 PARTITION OF app.events FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
INSERT INTO app.events VALUES (1, '2025-06-01', 'a'), (2, '2026-02-01', 'b'), (3, '2026-03-01', 'c');
CREATE MATERIALIZED VIEW app.event_counts AS
  SELECT kind, count(*) AS n FROM app.events GROUP BY kind;
-- An extension, a domain with a check, and a composite type, each used by a
-- column. dbmeta dumps them from v0.41.0 (W77).
CREATE EXTENSION IF NOT EXISTS citext;
CREATE DOMAIN app.email AS citext CHECK (VALUE LIKE '%@%');
CREATE TYPE app.address AS (street text, city text, zip integer);
CREATE TABLE app.contacts (
  contact_id integer PRIMARY KEY,
  email app.email NOT NULL,
  home app.address
);
INSERT INTO app.contacts VALUES (1, 'Ann@Example.com', ROW('1 Main St', 'Springfield', 12345)), (2, 'b@x', NULL);
