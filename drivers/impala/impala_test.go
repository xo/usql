package impala

import (
	"testing"

	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/stmt"
)

var impalaUrl, _ = dburl.Parse("impala://localhost:21050/")

func TestProcess(t *testing.T) {
	var process = func(sqlstr string) (string, bool) {
		prefix := stmt.FindPrefix(sqlstr, true, true, true)
		typ, sqlres, isQuery, err := drivers.Process(impalaUrl, prefix, sqlstr)
		if err != nil {
			t.Fatal(err)
		}
		if sqlres != sqlstr {
			t.Errorf("expected %q, got %q", sqlstr, sqlres)
		}
		return typ, isQuery
	}

	t.Run("SET", func(t *testing.T) {
		for sqlstr, isQuery := range map[string]bool{
			" SET foo=bar":           false,
			"SET ALL":                true,
			"--SET foo=bar\nSET":     true,
			"/* SET foo=bar */\nSET": true,
		} {
			t.Run(sqlstr, func(t *testing.T) {
				typ, isQueryRes := process(sqlstr)
				if isQuery != isQueryRes {
					t.Errorf("expected isQuery=%t, got %t", isQuery, isQueryRes)
				}
				if typ != "SET" {
					t.Errorf("expected typ SET, got %q", typ)
				}
			})
		}

	})

}
