package dameng

import (
	"reflect"
	"strings"
	"testing"

	"github.com/godoes/gorm-dameng/dm8"
	"github.com/xo/dburl"
)

// TestGenerateDSN verifies the external compatibility contract and native DM8 output.
func TestGenerateDSN(t *testing.T) {
	// tests cover the supported path, aliases, defaults, native SSL options, and rejected ambiguity.
	tests := []struct {
		// name identifies the compatibility scenario in test output.
		name string
		// input is the portable connection URL supplied to usql.
		input string
		// wantDSN is the native string passed to the DM8 driver.
		wantDSN string
		// wantSSLFilesPath checks the value parsed by the DM8 driver itself.
		wantSSLFilesPath string
		// wantErr is the expected validation error fragment.
		wantErr string
	}{
		{
			name:    "path schema and disabled SSL",
			input:   "dm://SYSDBA:pwd@127.0.0.1:5236/APPDB?sslmode=disable",
			wantDSN: "dm://SYSDBA:pwd@127.0.0.1:5236?schema=APPDB",
		},
		{
			name:             "long alias and native SSL",
			input:            "dameng://SYSDBA:pwd@db.example:5236?schema=APPDB&sslFilesPath=%2Fcerts",
			wantDSN:          "dm://SYSDBA:pwd@db.example:5236?schema=APPDB&sslFilesPath=/certs",
			wantSSLFilesPath: "/certs",
		},
		{
			name:    "dm8 alias",
			input:   "dm8://SYSDBA:pwd@localhost/APPDB",
			wantDSN: "dm://SYSDBA:pwd@localhost:5236?schema=APPDB",
		},
		{
			name:    "default port and explicit empty password",
			input:   "dm://SYSDBA@localhost/APPDB",
			wantDSN: "dm://SYSDBA:@localhost:5236?schema=APPDB",
		},
		{
			name:    "decoded password",
			input:   "dm://SYSDBA:p%40ss@localhost/APPDB",
			wantDSN: "dm://SYSDBA:p@ss@localhost:5236?schema=APPDB",
		},
		{
			name:    "conflicting schema",
			input:   "dm://SYSDBA:pwd@localhost/ONE?schema=TWO",
			wantErr: "both path and query",
		},
		{
			name:    "unsupported secure sslmode",
			input:   "dm://SYSDBA:pwd@localhost/APPDB?sslmode=verify-full",
			wantErr: "unsupported dameng sslmode",
		},
		{
			name:    "nested schema path",
			input:   "dm://SYSDBA:pwd@localhost/ONE/TWO",
			wantErr: "exactly one segment",
		},
		{
			name:    "IPv6 host",
			input:   "dm://SYSDBA:pwd@[::1]/APPDB",
			wantDSN: "dm://SYSDBA:pwd@[::1]:5236?schema=APPDB",
		},
		{
			name:    "native option containing separator",
			input:   "dm://SYSDBA:pwd@localhost/APPDB?sslFilesPath=%2Fcerts%26backup",
			wantErr: "cannot contain",
		},
	}

	// Execute every compatibility case through dburl.Parse, matching real usql behavior.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// parsedURL is the normalized connection definition returned to usql.
			parsedURL, err := dburl.Parse(test.input)
			// Error cases must fail with the agreed actionable message.
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("dburl.Parse() error = %v, want containing %q", err, test.wantErr)
				}
				return
			}
			// Successful cases must parse before their driver fields can be checked.
			if err != nil {
				t.Fatalf("dburl.Parse() unexpected error: %v", err)
			}
			// Every supported alias must resolve to the registered DM8 driver.
			if parsedURL.Driver != "dm" {
				t.Fatalf("Driver = %q, want %q", parsedURL.Driver, "dm")
			}
			// The generated DSN is the exact string passed to database/sql.
			if parsedURL.DSN != test.wantDSN {
				t.Fatalf("DSN = %q, want %q", parsedURL.DSN, test.wantDSN)
			}
			// Driver-level parsing is required only for native options whose escaping caused the regression.
			if test.wantSSLFilesPath == "" {
				return
			}
			// driverConnector is the DM8 driver's parsed representation of the generated DSN.
			driverConnector, err := (&dm8.DmDriver{}).OpenConnector(parsedURL.DSN)
			// A generated DSN must be accepted by the selected driver before inspecting its properties.
			if err != nil {
				t.Fatalf("OpenConnector() unexpected error: %v", err)
			}
			// connectorValue exposes the driver's parsed sslFilesPath for a black-box boundary assertion.
			connectorValue := reflect.ValueOf(driverConnector).Elem()
			// parsedSSLFilesPath is the native value that the driver will use to locate certificates.
			parsedSSLFilesPath := connectorValue.FieldByName("sslFilesPath").String()
			// The native path must remain decoded instead of containing URL escape bytes.
			if parsedSSLFilesPath != test.wantSSLFilesPath {
				t.Fatalf("sslFilesPath = %q, want %q", parsedSSLFilesPath, test.wantSSLFilesPath)
			}
		})
	}
}

// TestIsPasswordErr verifies interactive password prompting for English and Chinese driver errors.
func TestIsPasswordErr(t *testing.T) {
	// passwordError represents the localized authentication failure returned by DM8.
	passwordError := &testError{message: "用户名或密码错误"}
	// Authentication errors should allow usql to prompt once for a corrected password.
	if !isPasswordErr(passwordError) {
		t.Fatal("isPasswordErr() = false, want true")
	}
	// Network errors must not trigger a misleading password prompt.
	if isPasswordErr(&testError{message: "网络通信异常"}) {
		t.Fatal("isPasswordErr() = true for network error, want false")
	}
}

// testError supplies deterministic localized errors without requiring a live DM8 server.
type testError struct {
	// message is the exact driver-facing error text returned by Error.
	message string
}

// Error implements error for password classification tests.
func (err *testError) Error() string {
	return err.message
}
