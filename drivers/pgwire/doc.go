// Package pgwire gives the driver for a database that speaks the PostgreSQL
// protocol, such as Redshift or CockroachDB. It is the pgx driver when the
// build has pgx, and the libpq driver when it does not, so that the tags
// no_pgx and no_postgres leave pgx out of the binary. The package pgshared
// holds the code that the pgx and libpq drivers share.
//
// The build constraint of pgx.go is the one gen.go writes for the pgx driver
// in internal/pgx.go, and libpq.go has its negation. TestPgwireConstraints
// checks both.
package pgwire
