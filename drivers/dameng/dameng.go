// Package dameng defines and registers usql's Dameng DM8 driver.
//
// See: https://github.com/godoes/gorm-dameng
// Group: all
package dameng

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	_ "github.com/godoes/gorm-dameng/dm8" // DRIVER: dm
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

// init registers the DM8 driver's usql behavior. The dm, dm8, and dameng URL
// schemes and their DSN generation come from dburl.
func init() {
	// trailingSemicolonRE removes a final statement separator rejected by Oracle-compatible drivers.
	trailingSemicolonRE := regexp.MustCompile(`;?\s*$`)
	// endBlockRE identifies procedural blocks whose END separator must remain intact.
	endBlockRE := regexp.MustCompile(`(?i)\send\s*;\s*$`)

	// Register the usql behavior shared by every alias of the DM8 database/sql driver.
	drivers.Register("dm", drivers.Driver{
		AllowMultilineComments: true,
		LowerColumnNames:       true,
		Version:                version,
		User:                   currentUser,
		IsPasswordErr:          isPasswordErr,
		Process: func(_ *dburl.URL, prefix string, sqlstr string) (string, string, bool, error) {
			// Preserve the semicolon only when the SQL ends with a procedural END block.
			if !endBlockRE.MatchString(sqlstr) {
				sqlstr = trailingSemicolonRE.ReplaceAllString(sqlstr, "")
			}
			// queryType and query indicate how usql must dispatch the normalized statement.
			queryType, query := drivers.QueryExecType(prefix, sqlstr)
			return queryType, sqlstr, query, nil
		},
		NewMetadataReader: newMetadataReader,
		NewMetadataWriter: func(db drivers.DB, writer io.Writer, opts ...metadata.ReaderOption) metadata.Writer {
			// reader applies the same DM8 overrides used by completion and direct metadata calls.
			reader := newMetadataReader(db, opts...)
			// NewDefaultWriter exposes the shared \d, \dm, and \dp output behavior over the composed reader.
			return metadata.NewDefaultWriter(reader)(db, writer)
		},
		Copy: drivers.CopyWithInsert(func(position int) string {
			// DM8 uses Oracle-style numbered placeholders for each copied value.
			return fmt.Sprintf(":%d", position)
		}),
	})
}

// version returns the DM8 server banner shown by usql after connecting.
func version(ctx context.Context, db drivers.DB) (string, error) {
	// banner receives the first DM8 version banner exposed by the system view.
	var banner string
	// Return the query error so usql reports an unusable connection accurately.
	if err := db.QueryRowContext(ctx, `SELECT BANNER FROM SYS.V$VERSION WHERE ROWNUM = 1`).Scan(&banner); err != nil {
		return "", err
	}
	return banner, nil
}

// currentUser returns the authenticated DM8 account name.
func currentUser(ctx context.Context, db drivers.DB) (string, error) {
	// username receives the account reported by DM8's Oracle-compatible USER expression.
	var username string
	// Return the query error so prompts never display a guessed identity.
	if err := db.QueryRowContext(ctx, `SELECT USER FROM DUAL`).Scan(&username); err != nil {
		return "", err
	}
	return username, nil
}

// isPasswordErr reports whether reconnecting with an interactively supplied password may succeed.
func isPasswordErr(err error) bool {
	// message normalizes English and Chinese DM8 authentication errors for matching.
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "password") ||
		strings.Contains(message, "authentication") ||
		strings.Contains(message, "login") ||
		strings.Contains(message, "密码") ||
		strings.Contains(message, "口令") ||
		strings.Contains(message, "用户名")
}
