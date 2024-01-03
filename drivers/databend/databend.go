// Package databend defines and registers usql's Databend driver.
//
// See: https://github.com/datafuselabs/databend-go
package databend

import (
	"io"

	_ "github.com/datafuselabs/databend-go" // DRIVER
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
		UseColumnTypes:    true,
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
	})
}
