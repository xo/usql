# D6. Every setting has a section, and one way to set it

Status: Decided.

The maintainer decided on 2026-11-09 that every setting of usql has a full name,
SECTION.NAME, and that one model sets them all. W80 in
[BACKLOG.md](../BACKLOG.md) holds the plan and its steps.

## The model

- The sections are `usql`, `print`, `vars`, `charts`, `env`, `connections`
  and `aliases`. `usql` holds the settings of `\uset`, `print` those of
  `\pset`, `vars` the script variables of `\set`, `connections` the named
  connections of `\cset`, and `env` the variables of `\setenv`. `charts`
  holds the settings of `\chart`, which start with `charts.path`, because
  more of them are coming.
- `config.yaml` has one map for each section. `\config` sets and shows any
  setting by its full name, and `-O NAME=VALUE` sets one on the command line.
  `\pset`, `\uset`, `\set`, `\cset` and `\setenv` stay, and each sets its
  own section. `-v`, `-U`, `-N` and `-P` stay too.
- The settings apply in this order, and the last wins: the defaults, which
  read `PAGER` and `EDITOR`, then `config.yaml`, then the `USQL_*` variables,
  then the command line, then `init.sql`, then the session.
- Loading `config.yaml` prints nothing. Its errors go to standard error, and
  usql goes on.
- The settings of `env` reach only the commands that usql runs. The
  environment of usql does not change.
- Only `config.yaml` sets an alias. An alias cannot take the name of a meta
  command.

## Why

Before, each kind of setting had its own command, its own option, and its own
way into `config.yaml`, which read only `connections` and `charts_path`. A
setting of `\pset` could not be kept in the file. One name for each setting
lets the file, the command line and the session set the same thing in the same
words.

## What it does not do

- usql reads `config.yaml` with goccy/go-yaml, because viper turns keys into
  lower case, and the names of `vars`, `env` and `aliases` keep their case.
  cobra and viper stay for the command line until xo/ox can replace them.
- `\config` saves nothing to the file.
- An alias takes no arguments yet. How it would take them is an open question
  in [PLAN.md](../PLAN.md). Adding arguments later breaks no alias that
  exists.
