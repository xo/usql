package main

import (
	"maps"
	"slices"
	"testing"

	"github.com/xo/dbmeta"
	"github.com/xo/dburl"
	"github.com/xo/usql/drivers"
)

// noModel names each driver whose database dbmeta has no model for, and why.
// Its describe commands, such as \d, report that they are not supported. "no
// model yet" and "cannot have one" are different states, and a gap that is
// written down is one that can be counted. See step 4 of docs/DBMETA.md.
var noModel = map[string]string{
	"chai":       "no model yet",
	"cosmos":     "no model yet",
	"csvq":       "no model yet",
	"dynamodb":   "no model yet",
	"flightsql":  "no model yet",
	"maxcompute": "no model yet",
	"odbc":       "cannot have one: ODBC is a bridge to any database, and has no dialect of its own",
	"ots":        "no model yet",
	"pinot":      "no model yet",
	"voltdb":     "no model yet",
}

// TestEveryDriverReadsThroughDbmeta checks that each registered driver either
// reads its metadata through dbmeta or has an entry in noModel saying why not.
// A driver cannot pass by doing nothing, and an entry goes stale, and fails,
// once dbmeta has a model for it.
func TestEveryDriverReadsThroughDbmeta(t *testing.T) {
	t.Parallel()
	dialect := make(map[string]string)
	for _, scheme := range dburl.BaseSchemes() {
		dialect[scheme.Name] = scheme.Dialect
	}
	var read int
	for _, name := range slices.Sorted(maps.Keys(drivers.Available())) {
		_, has := dbmeta.Dialect(dialect[name]).Info()
		reason, exempt := noModel[name]
		switch {
		case has && exempt:
			t.Errorf("driver %s reads through dbmeta, and noModel still says %q. Remove its entry.", name, reason)
		case !has && !exempt:
			t.Errorf("driver %s has no dbmeta model for the dialect %q. Add it to noModel with the reason.", name, dialect[name])
		case has:
			read++
		}
	}
	if read == 0 {
		t.Error("no driver reads through dbmeta, so this guards nothing")
	}
}
