-- chart.sql makes a PostgreSQL database, fills a table, and draws the table
-- with \chart in the terminal.
--
-- Warning: do not run this against a server whose database usql_chart you
-- want to keep. The script drops that database and makes it again.
--
-- The script opens its connections with the libpq variables, which the
-- PostgreSQL driver of usql reads. It holds no host and no password.
--
-- 1. Set the variables for the server, for example:
--
--      export PGHOST=localhost PGPORT=5432 PGUSER=postgres PGPASSWORD=secret
--
-- 2. Start usql at a terminal that draws graphics, such as kitty, ghostty,
--    WezTerm or iTerm2:
--
--      usql pg:
--
-- 3. Run the script from the prompt:
--
--      \i _samples/chart.sql
--
-- Do not run it with usql -f. usql turns the terminal graphics off when it
-- reads a file given on the command line, and \chart then reports that the
-- terminal has no graphics.

-- the database of the user, on the server that the variables name
\c pg:

drop database if exists usql_chart;
create database usql_chart;

-- the new database, on the same server
\c pg:?dbname=usql_chart

create table sales (
  month date primary key,
  coffee integer not null,
  tea integer not null,
  cocoa integer not null
);

insert into sales (month, coffee, tea, cocoa) values
  ('2026-01-01', 820, 410, 390),
  ('2026-02-01', 790, 430, 350),
  ('2026-03-01', 860, 470, 280),
  ('2026-04-01', 910, 520, 190),
  ('2026-05-01', 980, 560, 120),
  ('2026-06-01', 1040, 610, 80),
  ('2026-07-01', 1100, 650, 60),
  ('2026-08-01', 1080, 640, 70),
  ('2026-09-01', 1000, 580, 140),
  ('2026-10-01', 930, 520, 230),
  ('2026-11-01', 870, 470, 330),
  ('2026-12-01', 850, 440, 420);

-- \chart runs the query and draws its result. The first column that is not
-- a number gives the labels of the x axis. Each column that is a number is
-- one series. \chart help lists the options.

-- a bar chart, which is the default when a column is not a number
select to_char(month, 'Mon') as month, coffee, tea, cocoa
from sales
order by sales.month
\chart title='Cups sold in 2026' subtitle='by month' type=bar

-- a line chart, in another size
select to_char(month, 'Mon') as month, coffee, tea, cocoa
from sales
order by sales.month
\chart title='Cups sold in 2026' type=line size=900x450
