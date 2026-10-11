package complete

import (
	"testing"

	"github.com/xo/usql/env"
	"github.com/xo/usql/text"
)

func TestCompleter(t *testing.T) {
	// a script variable that no case completes, so that completing one finds
	// names and none of them match
	_ = env.Vars().Set("zvar", "1")
	cases := []struct {
		name           string
		line           string
		start          int
		expSuggestions []string
		expLength      int
	}{
		{
			"Single SQL keyword, uppercase",
			"SEL",
			3,
			[]string{
				"ECT",
			},
			3,
		},
		{
			"Single SQL keyword, lowercase",
			"ex",
			2,
			[]string{
				"ec",
				"ecute",
				"plain",
			},
			2,
		},
		{
			"usql command",
			`\dt`,
			3,
			[]string{
				`+`,
				``,
				`S+`,
				`S`,
			},
			3,
		},
		{
			"files",
			`\i comp`,
			7,
			[]string{
				"leter.go",
				"leter_test.go",
			},
			4,
		},
		{
			"connections",
			`\c p`,
			4,
			[]string{
				"g://",
			},
			1,
		},
		{
			"3rd word",
			"SELECT * F",
			10,
			[]string{
				"ULL OUTER JOIN",
				"ROM",
				"ETCH",
			},
			1,
		},
		{
			"Selectables",
			"SELECT * FROM ",
			14,
			[]string{
				"main",
				"remote",
				"default",
				"system",
				"film",
				"factory",
			},
			0,
		},
		{
			"Namespaced with catalog",
			"SELECT * FROM remote.",
			21,
			[]string{
				"film",
				"factory",
			},
			7,
		},
		{
			"Namespaced with schema",
			"SELECT * FROM system.",
			21,
			[]string{
				"film",
				"factory",
			},
			7,
		},
		{
			"Namespaced with catalog.schema",
			"SELECT * FROM remote.default.f",
			30,
			[]string{
				"ilm",
				"actory",
			},
			16,
		},
		{
			"Attributes",
			"SELECT * FROM film WHERE ",
			25,
			[]string{
				"id",
				"name",
				"CASE",
				"AND",
				"OR",
				"WHEN",
				"THEN",
				"ELSE",
				"END",
			},
			0,
		},
		{
			"insert",
			"INS",
			3,
			[]string{
				"ERT",
			},
			3,
		},
		{
			"insert into",
			"INSERT IN",
			9,
			[]string{
				"TO",
			},
			2,
		},
		{
			"insert into table",
			"INSERT INTO fi",
			14,
			[]string{
				"lm",
			},
			2,
		},
		{
			"insert into table select from",
			"INSERT INTO film SE",
			19,
			[]string{
				"LECT",
			},
			2,
		},
		{
			"insert into table attrs",
			"INSERT INTO film (",
			18,
			[]string{
				"id",
				"name",
			},
			0,
		},
		{
			"insert into table values",
			"INSERT INTO film (a)",
			20,
			[]string{
				"SELECT",
				"TABLE",
				"VALUES",
				"OVERRIDING",
			},
			0,
		},
		{
			"update table set attrs",
			"UPDATE film SET ",
			16,
			[]string{
				"id",
				"name",
			},
			0,
		},
		{
			"update table set",
			"update film set name ",
			21,
			[]string{
				"=",
			},
			0,
		},
		{
			"variables",
			":a",
			2,
			[]string{},
			2,
		},
		{
			"type on create",
			"CREATE ",
			7,
			[]string{
				"SCHEMA",
				"DATABASE",
				"TABLE",
				"SEQUENCE",
				"VIEW",
				"TEMPORARY",
			},
			0,
		},
		{
			"brackets on create table",
			"CREATE TABLE p ",
			15,
			[]string{
				"(",
			},
			0,
		},
		{
			"type on alter",
			"ALTER ",
			6,
			[]string{
				"DATABASE",
				"FUNCTION",
				"INDEX",
				"SCHEMA",
				"SEQUENCE",
				"TABLE",
				"TRIGGER",
				"USER",
				"VIEW",
			},
			0,
		},
		{
			"type on drop",
			"DROP ",
			5,
			[]string{
				"DATABASE",
				"FUNCTION",
				"INDEX",
				"SCHEMA",
				"SEQUENCE",
				"TABLE",
				"TRIGGER",
				"USER",
				"VIEW",
			},
			0,
		},
		{
			"TABLE Selectables",
			"TABLE ",
			6,
			[]string{
				"main",
				"remote",
				"default",
				"system",
				"film",
				"factory",
			},
			0,
		},
		{
			"TABLE namespaced with catalog",
			"TABLE remote.",
			13,
			[]string{
				"film",
				"factory",
			},
			7,
		},
		{
			"TABLE namespaced with schema",
			"TABLE system.",
			13,
			[]string{
				"film",
				"factory",
			},
			7,
		},
		{
			"TABLE namespaced with catalog.schema",
			"TABLE remote.default.f",
			22,
			[]string{
				"ilm",
				"actory",
			},
			16,
		},
	}

	completer := New(WithSource(mockSource{}), WithConnStrings([]string{"pg://"}))
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			suggestions, length := completer.Do([]rune(test.line), test.start)
			// need at least 2 pairs of nested loops, one for what's missing, second for what's extra
			for _, exp := range test.expSuggestions {
				found := false
				for _, act := range suggestions {
					if string(act) == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Missing expected suggestion: %s", exp)
				}
			}
			for _, act := range suggestions {
				found := false
				for _, exp := range test.expSuggestions {
					if string(act) == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected suggestion: %s", string(act))
				}
			}
			if length != test.expLength {
				t.Errorf("Expected Do() to return length %d, got %d", test.expLength, length)
			}
		})
	}
}

// mockSource answers a fixed set of objects, the way a database would.
type mockSource struct{}

var _ Source = mockSource{}

func (mockSource) Objects(kind Kind, f Filter) ([]Object, error) {
	switch kind {
	case Catalogs:
		return []Object{{Name: "main"}, {Name: "remote"}}, nil
	case Schemas:
		return []Object{{Catalog: "main", Name: "default"}, {Catalog: "main", Name: "system"}}, nil
	case Tables:
		return []Object{
			{Catalog: f.Catalog, Schema: f.Schema, Name: "film"},
			{Catalog: f.Catalog, Schema: f.Schema, Name: "factory"},
		}, nil
	case Columns:
		if f.Parent == "film" {
			return []Object{{Name: "id"}, {Name: "name"}}, nil
		}
		return []Object{{Name: f.Catalog}, {Name: f.Schema}, {Name: f.Name}}, nil
	}
	return nil, text.ErrNotSupported
}
