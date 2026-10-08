-- roles and privileges
create role pd_owner login;
create role pd_reader;
create role pd_writer in role pd_reader;
alter role pd_reader set work_mem = '8MB';
create schema app authorization pd_owner;
create schema "Mixed Case";
comment on schema app is 'the app schema';
create extension if not exists pg_trgm;
create extension if not exists hstore;
-- types
create type app.mood as enum ('sad','ok','happy');
create type app.pair as (a int, b text);
create type app.span as range (subtype = int4);
create domain app.posint as int check (value > 0);
create domain app.email as text not null check (value like '%@%');
comment on type app.mood is 'a mood';
-- tables
create table app.account (
  id bigint generated always as identity primary key,
  email app.email unique,
  name text collate "C" not null,
  mood app.mood default 'ok',
  bal numeric(12,2) default 0 check (bal >= 0),
  tags text[] default '{}',
  meta jsonb,
  big text storage external,
  created timestamptz not null default now(),
  upper_name text generated always as (upper(name)) stored
) with (fillfactor = 70);
alter table app.account alter column name set statistics 500;
comment on table app.account is 'accounts';
comment on column app.account.email is 'the email';
create index account_meta_gin on app.account using gin (meta);
create index account_name_trgm on app.account using gin (name gin_trgm_ops);
create index account_partial on app.account (created desc) where bal > 100;
create unique index account_lower on app.account (lower(name)) include (bal);
create table app.orders (
  id serial primary key,
  account_id bigint not null references app.account(id) on delete cascade on update restrict deferrable initially deferred,
  amount app.posint,
  placed date default current_date,
  exclude using gist (int4range(id, id+1) with &&)
);
create table app.order_line (
  order_id int, line int, sku text, qty int,
  primary key (order_id, line),
  foreign key (order_id) references app.orders(id)
);
create table app.part (id int, region text, v int) partition by list (region);
create table app.part_eu partition of app.part for values in ('eu');
create table app.part_us partition of app.part for values in ('us','ca');
create table app.part_def partition of app.part default;
create table app.rng (d date, v int) partition by range (d);
create table app.rng_2026 partition of app.rng for values from ('2026-01-01') to ('2027-01-01');
create table app.base (id int, n text);
create table app.child1 (extra int) inherits (app.base);
create table app.child2 (more int) inherits (app.base, app.orders);
create unlogged table app.unl (x int);
create temp table pd_temp (x int);
create table "Mixed Case"."Odd Table" ("Odd Col" int primary key, "x" text);
create table app.noowner (x int);
alter table app.noowner owner to pd_owner;
create table app.rls (id int, owner_name text);
alter table app.rls enable row level security;
alter table app.rls force row level security;
create policy rls_p on app.rls for select to pd_reader using (owner_name = current_user);
create policy rls_w on app.rls for all to pd_writer using (true) with check (id > 0);
-- views
create view app.v_acct as select id, name from app.account;
create view app.v_secure with (security_barrier = true) as select * from app.v_acct where id > 0;
create view app.v_check as select * from app.v_acct where id > 0 with local check option;
create materialized view app.mv as select account_id, sum(amount) s from app.orders group by 1;
create unique index mv_idx on app.mv (account_id);
-- sequences
create sequence app.sq start 100 increment 5 minvalue 10 maxvalue 1000 cycle cache 3;
create sequence app.sq_owned; alter sequence app.sq_owned owner to pd_owner; alter sequence app.sq_owned owned by app.noowner.x;
-- functions, procedures, aggregates, operators
create function app.add(a int, b int default 1) returns int language sql immutable parallel safe as 'select a+b';
create function app.setof_t() returns setof app.account language sql stable as 'select * from app.account';
create function app.sec() returns text language plpgsql security definer set search_path = app as $$ begin return 'x'; end $$;
create function app.trg() returns trigger language plpgsql as $$ begin return new; end $$;
create procedure app.proc(x int) language plpgsql as $$ begin null; end $$;
create function app.leak(int) returns int language sql leakproof immutable as 'select $1';
create aggregate app.agg_sum(int) (sfunc = int4pl, stype = int, initcond = '0');
create function app.ieq(int,int) returns bool language sql immutable as 'select $1 = $2';
create operator app.=== (leftarg = int, rightarg = int, function = app.ieq, commutator = ===);
create cast (app.pair as text) with inout;
-- triggers, rules, event triggers
create trigger t_ins before insert on app.account for each row execute function app.trg();
create trigger t_upd after update of name on app.account for each row when (old.name is distinct from new.name) execute function app.trg();
create rule r_ins as on insert to app.noowner do also notify noowner;
create function app.evt() returns event_trigger language plpgsql as $$ begin null; end $$;
create event trigger e1 on ddl_command_end execute function app.evt();
-- extended statistics
create statistics app.st1 (dependencies) on id, name from app.account;
-- text search
create text search dictionary app.simple_d (template = simple, stopwords = english);
create text search configuration app.cfg (copy = english);
-- collation, conversion
create collation app.coll (provider = libc, locale = 'C');
-- foreign data
create foreign data wrapper pd_fdw;
create server pd_srv foreign data wrapper pd_fdw options (host 'x');
create user mapping for pd_owner server pd_srv options (user 'u');
create foreign table app.ft (id int, v text) server pd_srv options (table_name 't');
-- publication
create publication pd_pub for table app.account, app.orders;
-- large object
select lo_create(424242);
-- privileges
grant usage on schema app to pd_reader;
grant select on app.account, app.v_acct to pd_reader;
grant insert, update (name) on app.account to pd_writer;
grant execute on function app.add(int,int) to pd_reader;
alter default privileges in schema app grant select on tables to pd_reader;
alter default privileges for role pd_owner grant usage on sequences to pd_writer;
alter database pgdesc set search_path = public, app;
-- comments
comment on function app.add(int,int) is 'adds';
comment on view app.v_acct is 'a view';
comment on sequence app.sq is 'a sequence';
comment on index app.account_partial is 'partial';
comment on operator app.=== (int,int) is 'strict eq';
-- data, so sizes and rows are not zero
insert into app.account (email,name,meta) select 'u'||i||'@x', 'n'||i, '{"a":1}' from generate_series(1,500) i;
insert into app.orders (account_id, amount, placed) select (i%500)+1, 5, current_date from generate_series(1,200) i;
refresh materialized view app.mv;
analyze;
