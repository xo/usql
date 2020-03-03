package firebird

import (
	// DRIVER: firebirdsql
	_ "github.com/nakagami/firebirdsql"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("firebirdsql", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SELECT rdb$get_context('SYSTEM', 'ENGINE_VERSION') FROM rdb$database;`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "Firebird " + ver, nil
		},
	})
}
