package completer

import (
	"testing"

	"github.com/xo/usql/drivers/metadata"
)

func TestCompleter(t *testing.T) {
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
	}

	completer := NewDefaultCompleter(mockReader{})(nil)
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

type mockReader struct{}

var _ metadata.CatalogReader = &mockReader{}
var _ metadata.BasicReader = &mockReader{}

func (r mockReader) Catalogs(metadata.Filter) (*metadata.CatalogSet, error) {
	return metadata.NewCatalogSet([]metadata.Catalog{
		{
			Catalog: "main",
		},
		{
			Catalog: "remote",
		},
	}), nil
}

func (r mockReader) Schemas(metadata.Filter) (*metadata.SchemaSet, error) {
	return metadata.NewSchemaSet([]metadata.Schema{
		{
			Schema:  "default",
			Catalog: "main",
		},
		{
			Schema:  "system",
			Catalog: "main",
		},
	}), nil
}

func (r mockReader) Tables(f metadata.Filter) (*metadata.TableSet, error) {
	return metadata.NewTableSet([]metadata.Table{
		{
			Catalog: f.Catalog,
			Schema:  f.Schema,
			Name:    "film",
		},
		{
			Catalog: f.Catalog,
			Schema:  f.Schema,
			Name:    "factory",
		},
	}), nil
}

func (r mockReader) Columns(f metadata.Filter) (*metadata.ColumnSet, error) {
	if f.Parent == "film" {
		return metadata.NewColumnSet([]metadata.Column{
			{
				Name: "id",
			},
			{
				Name: "name",
			},
		}), nil
	}
	return metadata.NewColumnSet([]metadata.Column{
		{
			Name: f.Catalog,
		},
		{
			Name: f.Schema,
		},
		{
			Name: f.Name,
		},
	}), nil
}
