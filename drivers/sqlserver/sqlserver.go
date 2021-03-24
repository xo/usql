// Package sqlserver defines and registers usql's Microsoft SQL Server driver.
//
// See: https://github.com/denisenkom/go-mssqldb
package sqlserver

import (
	"io"
	"strconv"
	"strings"

	sqlserver "github.com/denisenkom/go-mssqldb" // DRIVER: sqlserver
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

func init() {
	newReader := func(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
		ir := infos.New(
			infos.WithIndexes(false),
			infos.WithSequences(false),
			infos.WithCustomColumns(map[infos.ColumnName]string{
				infos.FunctionsSecurityType: "''",
			}),
			infos.WithSystemSchemas([]string{
				"db_accessadmin",
				"db_backupoperator",
				"db_datareader",
				"db_datawriter",
				"db_ddladmin",
				"db_denydatareader",
				"db_denydatawriter",
				"db_owner",
				"db_securityadmin",
				"INFORMATION_SCHEMA",
				"sys",
			}),
		)(db, opts...)
		mr := &metaReader{
			LoggingReader: metadata.NewLoggingReader(db, opts...),
		}
		return metadata.NewPluginReader(ir, mr)
	}
	drivers.Register("sqlserver", drivers.Driver{
		AllowMultilineComments:  true,
		RequirePreviousPassword: true,
		LexerName:               "tsql",
		Version: func(db drivers.DB) (string, error) {
			var ver, level, edition string
			err := db.QueryRow(
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
		NewMetadataReader: newReader,
		NewMetadataWriter: func(db drivers.DB, w io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			return metadata.NewDefaultWriter(newReader(db, opts...))(db, w)
		},
	})
}
