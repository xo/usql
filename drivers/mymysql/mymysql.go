// Package mymysql defines and registers usql's MySQL MyMySQL driver.
//
// See: https://github.com/ziutek/mymysql
package mymysql

import (
	"io"
	"strconv"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	mymeta "github.com/xo/usql/drivers/metadata/mysql"
	_ "github.com/ziutek/mymysql/godrv" // DRIVER
	"github.com/ziutek/mymysql/mysql"
)

func init() {
	drivers.Register("mymysql", drivers.Driver{
		AllowMultilineComments: true,
		AllowHashComments:      true,
		LexerName:              "mysql",
		UseColumnTypes:         true,
		Err: func(err error) (string, string) {
			if e, ok := err.(*mysql.Error); ok {
				return strconv.Itoa(int(e.Code)), string(e.Msg)
			}
			return "", err.Error()
		},
		IsPasswordErr: func(err error) bool {
			if e, ok := err.(*mysql.Error); ok {
				return e.Code == mysql.ER_ACCESS_DENIED_ERROR
			}
			return false
		},
		NewMetadataReader: mymeta.NewReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(mymeta.NewReader(db, opts...))(db, w)
		},
		Copy:         drivers.CopyWithInsert(func(int) string { return "?" }),
		NewCompleter: mymeta.NewCompleter,
	})
}
