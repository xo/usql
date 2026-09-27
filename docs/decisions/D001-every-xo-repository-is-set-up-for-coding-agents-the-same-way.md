# D1. Every xo repository is set up for coding agents the same way

Status: Decided.

Ken decided on 2026-09-27 that every repository in the `xo` namespace is set
up for coding agents the same way. dbmeta records the standard as its D110.
This decision adopts it for usql.

## The files

In the root:

- `README.md`, for a person who finds the project.
- `AGENTS.md`, which holds the rules for a coding agent. Codex and the other
  agents read this file.
- `CLAUDE.md`, which holds one line, `@AGENTS.md`. Claude Code reads
  `CLAUDE.md`, and the line imports `AGENTS.md`, so that every agent reads the
  same rules. It is a file and not a symbolic link, because a Windows checkout
  writes a link as a small text file.
- `CONTRIBUTING.md`, for a person who changes the project. It has an Agent
  skills section with the command that installs them.
- `skills-lock.json`, which names the source of each skill.
- `.gitignore`, which ignores `.claude/settings.local.json`, the Claude Code
  permissions of one person.
- `.gitattributes`, which holds `* text=auto eol=lf`.

In `docs/`:

- `PLAN.md`, which holds the plan and the open questions for Ken.
- `decisions/`, one file per decision, with an index. D2 says why.
- `BACKLOG.md`, which holds the work that is known and not done.

## The skills

The repository commits `simple-english` and `go-pedantry`. Each is an
ordinary folder in `.agents/skills/<name>` and in `.claude/skills/<name>`,
installed with `--copy`. `TestSkillsAreCopies` fails on a link, on a missing
copy and on two copies that differ. Ken installs the skills himself.

## The standing rules

`AGENTS.md` opens with three rules that are the same in every repository:

1. Stage changes for review. Commit and push only when Ken says so.
2. Load `simple-english` before writing any text that a person reads: project
   documentation, a code comment, an error message or a commit message.
3. Load `go-pedantry` before writing or reviewing Go code. A rule of the
   project wins where the two conflict.

Until now each agent learned the first rule from Ken and kept it in its own
memory, so an agent new to the repository did not know it.

## What changed in usql

`CLAUDE.md` moved to `AGENTS.md`, which now opens with the standing rules, and
`CLAUDE.md` holds the import. `TestClaudeImportsAgents` checks it. The root
holds four documents, not three, and `TestTheRootHoldsFourDocuments` checks
that. `.gitattributes`, `docs/PLAN.md` and this folder are new.

usql had no decision log before, so this is its first decision. `dburl` and
`dbmeta` number their decisions the same way. A reference to one of theirs
names the repository, as in "dburl's D22", and a bare number means usql's own.
