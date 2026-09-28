// Package neo4j defines and registers usql's Neo4j driver.
//
// Pre-flight, checked on 2026-09-29:
//
//   - Maintained: github.com/xo/dbimp v0.4.0 holds the driver, which v0.3.0
//     added on 2026-09-28. The repository is in the same family as usql.
//   - Global state: the package's init calls sql.Register and nothing else.
//   - Links: pure Go, over the Query API of HTTP.
//   - Tested: the vendor's image, which dbmeta's dbrun starts as
//     neo4j-2026.09.0.
//
// Metadata: none. All metadata is moving into dbmeta, so usql writes no
// reader for Neo4j (D3).
//
// See: https://github.com/xo/dbimp
// Group: most
package neo4j

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/xo/dbimp/neo4j" // DRIVER
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("neo4j", drivers.Driver{
		AllowMultilineComments: true,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// Neo4j sends no count of rows affected, and a Cypher statement
			// such as MATCH is not one that usql knows returns rows, so each
			// statement is run as a query
			typ, _ := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, true, nil
		},
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver, edition string
			err := db.QueryRowContext(ctx, `CALL dbms.components() YIELD name, versions, edition `+
				`WHERE name = 'Neo4j Kernel' RETURN versions[0], edition`).Scan(&ver, &edition)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return "Neo4j <unknown>", nil
			case err != nil:
				return "", err
			}
			return "Neo4j " + ver + " " + edition, nil
		},
		Err: func(err error) (string, string) {
			// a response can hold several errors, and only a single one is
			// reported by its code
			var re *neo4j.ResponseError
			if errors.As(err, &re) && len(re.Errs) == 1 {
				return re.Errs[0].Code, re.Errs[0].Message
			}
			return "", strings.TrimPrefix(err.Error(), "neo4j: ")
		},
		IsPasswordErr: func(err error) bool {
			var e neo4j.Error
			return errors.As(err, &e) && e.Code == "Neo.ClientError.Security.Unauthorized"
		},
	})
}
