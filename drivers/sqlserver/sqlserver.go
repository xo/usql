// Package sqlserver defines and registers usql's Microsoft SQL Server driver.
//
// See: https://github.com/denisenkom/go-mssqldb
// Group: base
package sqlserver

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	sqlserver "github.com/denisenkom/go-mssqldb" // DRIVER
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

func init() {
	drivers.Register("sqlserver", drivers.Driver{
		AllowMultilineComments:  true,
		RequirePreviousPassword: true,
		LexerName:               "tsql",
		Version: func(ctx context.Context, db drivers.DB) (string, error) {
			var ver, level, edition string
			err := db.QueryRowContext(
				ctx,
				`SELECT SERVERPROPERTY('productversion'), SERVERPROPERTY ('productlevel'), SERVERPROPERTY ('edition')`,
			).Scan(&ver, &level, &edition)
			if err != nil {
				return "", err
			}
			return "Microsoft SQL Server " + ver + ", " + level + ", " + edition, nil
		},
		ChangePassword: func(db drivers.DB, user, newpw, oldpw string) error {
			_, err := db.Exec(`ALTER LOGIN ` + user + ` WITH password = '` + newpw + `' old_password = '` + oldpw + `'`)
			return err
		},
		Err: func(err error) (string, string) {
			if e, ok := err.(sqlserver.Error); ok {
				return strconv.Itoa(int(e.Number)), e.Message
			}
			msg := err.Error()
			if i := strings.LastIndex(msg, "sqlserver:"); i != -1 {
				msg = msg[i:]
			}
			return "", msg
		},
		IsPasswordErr: func(err error) bool {
			return strings.Contains(err.Error(), "Login failed for")
		},
		NewMetadataReader: NewReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(NewReader(db, opts...))(db, w)
		},
		Copy: drivers.CopyWithInsert(placeholder),
	})
}

func placeholder(n int) string {
	return fmt.Sprintf("@p%d", n)
}
