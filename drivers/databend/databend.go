// Package databend defines and registers usql's Databend driver.
//
// The driver is dbimp's, which replaced github.com/datafuselabs/databend-go
// in dbimp v0.6.0. It refuses the keys of databend-go, such as sslmode,
// tenant and warehouse.
//
// See: https://github.com/xo/dbimp
// Group: most
package databend

import (
	"database/sql"
	"errors"
	"io"

	"github.com/xo/dbimp"
	_ "github.com/xo/dbimp/databend" // DRIVER
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	newReader := infos.New(
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithCustomClauses(map[infos.ClauseName]string{
			infos.SequenceColumnsIncrement: "''",
		}),
		infos.WithFunctions(false),
		infos.WithIndexes(false),
		infos.WithConstraints(false),
		infos.WithColumnPrivileges(false),
	)
	drivers.Register("databend", drivers.Driver{
		UseColumnTypes: true,
		// a statement such as CREATE TABLE sends no count of rows affected,
		// which the driver reports as ErrNotSupported (dbimp D125)
		RowsAffected: func(res sql.Result) (int64, error) {
			n, err := res.RowsAffected()
			if errors.Is(err, dbimp.ErrNotSupported) {
				return 0, nil
			}
			return n, err
		},
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
	})
}
