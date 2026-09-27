# D2. A large project keeps one file per decision

Status: Decided.

Ken decided on 2026-09-27 that a large project keeps each decision in a file
of its own, because a markdown file many thousands of lines long is not
readable. dbmeta records it as its D111, and names usql among the large
projects.

## The layout

Each decision is `docs/decisions/D<nnn>-<title>.md`, with the number in three
digits so that a listing sorts in order. The file opens with its number and
title, a blank line, and its status:

    # D1. Every xo repository is set up for coding agents the same way

    Status: Decided.

`docs/decisions/README.md` is the index, and GitHub shows it as the page of
the folder. `docs/PLAN.md` keeps the plan and the open questions.

A decision that changes an earlier one says so in its status, as "Amends D1",
and the earlier one says it back, as "Amended by D3". Each file opens with its
status, so a reader sees an amendment before anything else.

## The tests

`TestTheDecisionIndexIsComplete` checks every row of the index against the
file, title and status alike, and prints the row to add.
`TestAnAmendmentPointsBothWays` fails unless both decisions name each other in
their status. Both read the folder, and the file names and headings must
follow the layout above.
