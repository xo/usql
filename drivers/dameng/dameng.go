// Package dameng defines and registers usql's Dameng DM8 driver.
//
// See: https://github.com/godoes/gorm-dameng
// Group: most
package dameng

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"

	_ "github.com/godoes/gorm-dameng/dm8" // DRIVER: dm
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
)

const (
	// defaultPort is the standard DM8 TCP port.
	defaultPort = "5236"
	// schemaQueryKey is the DM8 driver's default-schema property.
	schemaQueryKey = "schema"
	// sslModeQueryKey is the PostgreSQL-style compatibility option accepted by usql.
	sslModeQueryKey = "sslmode"
)

// init registers the DM8 URL scheme and its usql behavior.
func init() {
	// trailingSemicolonRE removes a final statement separator rejected by Oracle-compatible drivers.
	trailingSemicolonRE := regexp.MustCompile(`;?\s*$`)
	// endBlockRE identifies procedural blocks whose END separator must remain intact.
	endBlockRE := regexp.MustCompile(`(?i)\send\s*;\s*$`)

	// Register the fork-local URL contract before usql parses any dm, dm8, or dameng connection.
	dburl.Register(dburl.Scheme{
		Driver:    "dm",
		Generator: generateDSN,
		Transport: dburl.TransportTCP,
		Aliases:   []string{"dameng", "dm8"},
	})
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

// generateDSN converts usql's URL contract into the DSN expected by the DM8 Go driver.
func generateDSN(connectionURL *dburl.URL) (string, string, error) {
	// query holds a mutable copy of the user-provided driver options.
	query := connectionURL.Query()
	// pathSchema treats the URL path as the caller's default schema.
	pathSchema := strings.Trim(connectionURL.Path, "/")
	// configuredSchema preserves an explicitly supplied native DM8 schema option.
	configuredSchema := query.Get(schemaQueryKey)

	// Reject nested paths because DM8 accepts one default schema, not a path hierarchy.
	if strings.Contains(pathSchema, "/") {
		return "", "", fmt.Errorf("dameng schema path must contain exactly one segment")
	}
	// Reject conflicting schema declarations instead of silently selecting one.
	if pathSchema != "" && configuredSchema != "" && !strings.EqualFold(pathSchema, configuredSchema) {
		return "", "", fmt.Errorf("dameng schema is specified by both path and query with different values")
	}
	// Translate the portable path form only when no native schema option already exists.
	if pathSchema != "" && configuredSchema == "" {
		query.Set(schemaQueryKey, pathSchema)
	}

	// sslMode is the normalized compatibility value from existing dbhub-style DSNs.
	sslMode := strings.ToLower(strings.TrimSpace(query.Get(sslModeQueryKey)))
	// Accept only the explicitly agreed compatibility mode; secure modes require native DM8 parameters.
	switch sslMode {
	// An omitted compatibility option leaves native DM8 SSL parameters untouched.
	case "":
	// The dbhub-style disabled mode maps to the DM8 driver's non-SSL default.
	case "disable":
		query.Del(sslModeQueryKey)
	// Other PostgreSQL-style modes cannot be translated without weakening their meaning.
	default:
		return "", "", fmt.Errorf("unsupported dameng sslmode %q; use sslFilesPath, sslCertPath, and sslKeyPath", sslMode)
	}

	// host defaults to localhost for consistency with other dburl schemes.
	host := connectionURL.Hostname()
	// Supply the DM8 host default only when the caller omitted it.
	if host == "" {
		host = "localhost"
	}
	// port defaults to DM8's standard listener port.
	port := connectionURL.Port()
	// Supply the standard port because the Go driver requires a host:port pair.
	if port == "" {
		port = defaultPort
	}

	// username and password are rebuilt with an explicit password separator for interactive prompting.
	username, password := "", ""
	// Read credentials only when the parsed URL includes user information.
	if connectionURL.User != nil {
		username = connectionURL.User.Username()
		password, _ = connectionURL.User.Password()
	}
	// Reject credential delimiters that the DM8 driver's non-URL parser cannot represent safely.
	if strings.ContainsAny(username, ":?") || strings.Contains(password, "?") {
		return "", "", fmt.Errorf("dameng username cannot contain ':' or '?', and password cannot contain '?'")
	}
	// rawQuery preserves decoded native DM8 parameter values such as filesystem paths.
	rawQuery, err := encodeDriverQuery(query)
	// Return validation failures before constructing a DSN that the driver would misparse.
	if err != nil {
		return "", "", err
	}
	// driverDSN is the path-free string understood by github.com/godoes/gorm-dameng/dm8.
	driverDSN := "dm://" + username + ":" + password + "@" + net.JoinHostPort(host, port)
	// Append native options only when at least one option remains after compatibility translation.
	if rawQuery != "" {
		driverDSN += "?" + rawQuery
	}
	return driverDSN, "", nil
}

// encodeDriverQuery serializes standard URL query values for the DM8 driver's raw parser.
func encodeDriverQuery(query url.Values) (string, error) {
	// Inspect every key and value because the DM8 driver splits options without URL decoding.
	for key, values := range query {
		// Reject key delimiters that would change the driver's option boundaries.
		if strings.ContainsAny(key, "&=?") {
			return "", fmt.Errorf("dameng option name %q cannot contain '&', '=', or '?'", key)
		}
		// Inspect every repeated value associated with the current native option.
		for _, value := range values {
			// Reject value delimiters that cannot be represented by the driver's query grammar.
			if strings.ContainsAny(value, "&?") {
				return "", fmt.Errorf("dameng option %q cannot contain '&' or '?'", key)
			}
		}
	}
	// rawQuery decodes the standard encoder's escapes after it supplies stable key ordering.
	rawQuery, err := url.QueryUnescape(query.Encode())
	// QueryUnescape should only fail for malformed escapes introduced outside url.Values.
	if err != nil {
		return "", fmt.Errorf("invalid dameng options: %w", err)
	}
	return rawQuery, nil
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
