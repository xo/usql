# Decisions

Every decision this project made is a file in this folder, one per decision,
named by its number and its title. This table is the index. Find the number
here, then open the file.

Each file opens with its status. "Decided" means Ken chose it. "Proposed"
means an agent or a peer session suggested it and Ken has not confirmed it.
"Open" means nobody has chosen yet. A decision that changes an earlier one
says so in its status, as "Amends D1", and the earlier one says it back, as
"Amended by D3". Read the status before the decision.

A new decision gets the next number and a file of its own. Add its row here.
`TestTheDecisionIndexIsComplete` fails when a decision has no row or a row is
wrong, and it prints the row to add. D2 holds the layout.

`dburl` and `dbmeta` keep their own decision logs with the same numbering. A
reference to one of theirs names the repository, as in "dburl's D22".

| # | Decision | Status |
| --- | --- | --- |
| [D1](D001-every-xo-repository-is-set-up-for-coding-agents-the-same-way.md) | Every xo repository is set up for coding agents the same way | Decided |
| [D2](D002-a-large-project-keeps-one-file-per-decision.md) | A large project keeps one file per decision | Decided |
