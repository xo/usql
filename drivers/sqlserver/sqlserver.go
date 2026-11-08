// Package sqlserver defines and registers usql's Microsoft SQL Server driver.
//
// See: https://github.com/microsoft/go-mssqldb
package sqlserver

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	sqlserver "github.com/microsoft/go-mssqldb" // DRIVER
	"github.com/xo/usql/drivers"

	// needed for azuresql authentication, named pipes, and shared memory transport protocols
	_ "github.com/microsoft/go-mssqldb/azuread"
	_ "github.com/microsoft/go-mssqldb/namedpipe"
	_ "github.com/microsoft/go-mssqldb/sharedmemory"
)

func init() {
	drivers.Register("sqlserver", drivers.Driver{
		/*
			// NOTE: this has been commented out, as it is not necessary. if
			// NOTE: the azuread.DriverName is changed from `azuresql`, then
			// NOTE: this func will be necessary as dburl will never import non
			// NOTE: stdlib package. as is, dburl.Open will handle the call
			// NOTE: to sql.Open and will pass the `azuresql` driver name
			Open: func(_ context.Context, u *dburl.URL, _, _ func() io.Writer) (func(string, string) (*sql.DB, error), error) {
				return func(_ string, params string) (*sql.DB, error) {
					driver := "sqlserver"
					switch {
					case u.Query().Has("fedauth"),
						strings.Contains(strings.ToLower(u.OriginalScheme), "azuresql"):
						driver = azuread.DriverName
					}
					return sql.Open(driver, params)
				}, nil
			},
		*/
		ColumnTypes: func(col *sql.ColumnType) (interface{}, error) {
			switch col.DatabaseTypeName() {
			case "UNIQUEIDENTIFIER":
				if nullable, ok := col.Nullable(); ok && nullable {
					return new(sql.Null[uniqueIdentifier]), nil
				}
				return new(uniqueIdentifier), nil
			case "REAL":
				// go-mssqldb widens a real to a float64, which then prints
				// 4.9 as 4.900000095367432, and a float32 prints as sqlcmd does
				if nullable, ok := col.Nullable(); ok && nullable {
					return new(sql.Null[float32]), nil
				}
				return new(float32), nil
			}
			return new(interface{}), nil
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
		// sqlcmd writes each message of the batch as Msg 208, Level 16,
		// State 1, Server srv, Line 1, and the message on the next line
		ErrVerbose: func(err error, _ string) string {
			var e sqlserver.Error
			if !errors.As(err, &e) {
				return ""
			}
			all := e.All
			if len(all) == 0 {
				all = []sqlserver.Error{e}
			}
			var lines []string
			for _, m := range all {
				s := fmt.Sprintf("Msg %d, Level %d, State %d", m.Number, m.Class, m.State)
				if m.ServerName != "" {
					s += ", Server " + m.ServerName
				}
				if m.ProcName != "" {
					s += ", Procedure " + m.ProcName
				}
				s += fmt.Sprintf(", Line %d", m.LineNo)
				lines = append(lines, s, m.Message)
			}
			return strings.Join(lines, "\n")
		},
		IsPasswordErr: func(err error) bool {
			return strings.Contains(err.Error(), "Login failed for")
		},
	})
}

// uniqueIdentifier is a uniqueidentifier of SQL Server that prints as a uuid of
// PostgreSQL does, in lower case, so that a column of UUIDs prints the same on
// both. go-mssqldb's UniqueIdentifier reads the bytes in the order that SQL
// Server stores them, and prints in upper case, as SQL Server's tools do.
type uniqueIdentifier struct {
	sqlserver.UniqueIdentifier
}

// String satisfies the fmt.Stringer interface, which tblfmt prints.
func (u uniqueIdentifier) String() string {
	return strings.ToLower(u.UniqueIdentifier.String())
}
