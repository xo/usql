// Package snowflake defines and registers usql's Snowflake driver.
//
// See: https://github.com/snowflakedb/gosnowflake
package snowflake

import (
	"io"
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/snowflakedb/gosnowflake" // DRIVER: snowflake
	"github.com/xo/tblfmt"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
	"github.com/xo/usql/env"
)

func init() {
	r := logrus.New()
	r.Out, r.Level = ioutil.Discard, logrus.PanicLevel
	var l gosnowflake.SFLogger = &logger{r}
	gosnowflake.SetLogger(&l)
	newReader := infos.New(
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithFunctions(false),
		infos.WithIndexes(false),
	)
	drivers.Register("snowflake", drivers.Driver{
		Err: func(err error) (string, string) {
			if e, ok := err.(*gosnowflake.SnowflakeError); ok {
				return strconv.Itoa(e.Number), e.Message
			}
			return "", err.Error()
		},
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			writerOpts := []metadata.WriterOption{
				metadata.WithListAllDbs(func(pattern string, verbose bool) error {
					return listAllDbs(db, w, pattern, verbose)
				}),
			}
			return metadata.NewDefaultWriter(newReader(db, opts...), writerOpts...)(db, w)
		},
	})
}

// logger is an empty logger.
type logger struct {
	*logrus.Logger
}

func (*logger) SetLogLevel(string) error { return nil }

func listAllDbs(db drivers.DB, w io.Writer, pattern string, verbose bool) error {
	rows, err := db.Query("SHOW databases")
	if err != nil {
		return err
	}
	defer rows.Close()

	params := env.Pall()
	params["title"] = "List of databases"
	return tblfmt.EncodeAll(w, rows, params)
}
