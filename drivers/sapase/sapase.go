package sapase

import (
	// DRIVER: tds
	_ "github.com/thda/tds"

	"github.com/xo/usql/drivers"
)

func init() {
	drivers.Register("tds", drivers.Driver{
		AllowMultilineComments: true,
		Version: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SELECT @@version`).Scan(&ver)
			if err != nil {
				return "", err
			}
			return "SAP ASE " + ver, nil
		},
	})
}
