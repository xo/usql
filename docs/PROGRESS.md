# Progress

This file says where the work in progress stands, so that a session that
stops can be picked up again. It is about W36 in
[BACKLOG.md](BACKLOG.md). Update it at each checkpoint.

## Work modes

Ken's machine has crashed from resource exhaustion. Work in a low-resource
mode:

- Build one package at a time. Do not run the whole test suite, and run no
  more than one test package at once.
- Start no containers and no background jobs unless Ken asks.
- Commit a checkpoint before a long step, on `main`, as a commit with a
  proper message. Ken wants no work branches and no WIP commits. Do not
  push.

## What W36 changes

usql reads its input through rline and the usql grammar of transit, in place
of gohxs/readline, usql's own `rline` package and the `stmt` package.

- `feed` is new. It parses input with the usql grammar, holds psql's query
  buffer, says when an entry typed in the editor is finished, feeds
  statements and meta commands one at a time from the editor or from lines,
  and expands variables when a statement runs. Its tests pass.
- `metacmd` reads the arguments of a meta command from the grammar's parts.
  `\g` reads its option list from the grammar. `\e`, `\p`, `\w` and `\r`
  work on the feed buffer. It builds.
- `handler` reads through `feed` and talks to the user through
  `*rline.Prompt`. Its loop is new. The chroma highlighting is gone. It
  builds.
- `drivers` has `Language(u)` in place of the stmt options, and no chroma
  lexer. A driver's own Version wins over dbmeta's, for InfluxDB's release
  from GET /ping (dbmeta's D166).
- `metacmd/complete`, which was `drivers/completer`, is an `rline.Completer`. It adds candidates with
  `AddCandidate`, because `Completion.Add` deletes nothing before it.
- `styles` is removed.

## What is left

1. Done on 2026-10-02: run.go and main.go use rline and feed, `rline/` and
   `stmt/` are deleted, chroma and gohxs/readline are out of `go.mod`, and
   the handler, metacmd and feed tests pass.
2. Checked by hand on SQLite, piped: statements across lines, `\set` and
   variables, `\g` for a statement with no `;`, `\echo`, `\i` with a
   variable that it sets, `\prompt`, `\p` and `\r`. Checked at a
   pseudo-terminal: an entry over two rows, variables and `\q`.
3. Checked on 2026-10-02: Ctrl-C gives up an open entry, `\e` runs the
   edited text and goes on, and `\c` from SQLite to DuckDB parses what
   follows with DuckDB's forms. The history is not saved: see below.
4. TestCLI passes: the nine sessions of `testdata/cli` were re-recorded on 2026-10-11,
   with no color of highlighting, and `line-editing` was rewritten, because Enter
   on a finished entry runs it wherever the cursor is, and rline keeps an open
   entry as rows, where readline ran its line.

Open points that are known:

- `%R` shows only `=` or `-`, because rline draws one prompt for an entry.
  A prompt for each row waits on rline's backlog item 33.
- The syntax_hl variables do nothing until transit's SQL grammars come.
- rline writes `\x1b[m` before the first text when color is on and the
  output is not a terminal. run.go turns color off unless both input and
  output are a terminal. Tell rline: its session was not running on
  2026-10-02.
- At a pseudo-terminal under `script`, the first keys typed after start are
  lost, and later keys are not. rline may ask the terminal a question when
  it starts and read the answer from the input. Tell rline. It may also lose
  type-ahead at a real terminal.
- The history is never saved for a user who has a history file from
  gohxs/readline. rline's format escapes a backslash, so the first line such
  as `\q` stops the read, and rline then refuses to save, to keep the rest of
  the file. It ignores the error at start. This is for Ken: migrate the file,
  use a new file name, or have rline read the old format.
- dbmeta gives the version of every database that it has a model for, and the
  Version hooks that remain are InfluxDB's (by Ken's decision, dbmeta's D166)
  and csvq's, for which dbmeta has no model. The describe commands read dbmeta
  through `metacmd/describe` (W46 holds what still differs from psql).

## How to build

usql requires dbmeta v0.15.0, transit v0.4.0 with its usql grammar v0.4.0, and
rline v0.1.0 in `go.mod`. It builds with plain modules and no `go.work`, and
`go mod tidy` leaves `go.mod` as it is. A workspace is still how to build
against a local checkout of a sibling that has not tagged a change yet: write a
`go.work` that uses `.` and the checkouts, and do not commit it. If the workspace
loads two modules that provide `github.com/google/flatbuffers/go`, a `replace`
to an empty module, a directory with only a `go.mod` that reads `module
github.com/google/flatbuffers/go` and `go 1.21`, ends the conflict.

The recordings in `drivers/testdata` were removed with the tests against
databases. W37 brings them back.
