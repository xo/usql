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
4. TestCLI fails in all nine sessions, because the recordings hold the
   colors of the old highlighting. Ken fixes the recordings himself. The
   session line-editing also runs into its limit of 30 seconds, which may be
   the fault below.

Open points that are known:

- `%R` shows only `=` or `-`, because rline draws one prompt for an entry.
  A prompt for each row waits on rline's backlog item 33.
- The SYNTAX_HL variables do nothing until transit's SQL grammars come.
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
- dbmeta v0.1.0 holds DuckDB's dollar quotes, and the `Query.Each`, `Args`,
  `Dialect.Pattern` and `Open` that `metacmd/describe` uses (dbmeta's D169).
- Do not run `go mod tidy`, which adds rline before it has a tag.

## How to build

usql requires dbmeta v0.1.0, transit v0.2.0 and transit's usql grammar
v0.2.0 in `go.mod`. It builds against a local checkout of rline, which has
no tag. `go.work` is not committed. It reads:

    go 1.27.1

    use (
        .
        ../rline
    )

    replace github.com/google/flatbuffers/go => /path/to/flatbuffers-go-empty

The replace points at an empty module, a directory with only a `go.mod` that
reads `module github.com/google/flatbuffers/go` and `go 1.21`. Workspace mode
loads two modules that both provide that package, and the empty one ends the
conflict. A crash can delete it if it is under a temporary directory.

The rline checkout was at bef86da on 2026-10-02. transit's D108 made the usql
grammar one language that takes the lexical forms of the dialect as options,
in place of six packages.

The recordings in `testdata/cli` are changed and not committed, by Ken's
decision. Keep them out of every commit. The recordings in `drivers/testdata`
were removed with the tests against databases. W37 brings them back.
