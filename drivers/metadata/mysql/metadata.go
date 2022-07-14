package mysql

import (
	"time"

	"github.com/gohxs/readline"
	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/completer"
	"github.com/xo/usql/drivers/metadata"
	infos "github.com/xo/usql/drivers/metadata/informationschema"
)

var (
	// NewReader for MySQL databases
	NewReader = infos.New(
		infos.WithPlaceholder(func(int) string { return "?" }),
		infos.WithSequences(false),
		infos.WithCheckConstraints(false),
		infos.WithCustomClauses(map[infos.ClauseName]string{
			infos.ColumnsDataType:                 "column_type",
			infos.ColumnsNumericPrecRadix:         "10",
			infos.FunctionColumnsNumericPrecRadix: "10",
			infos.ConstraintIsDeferrable:          "''",
			infos.ConstraintInitiallyDeferred:     "''",
			infos.ConstraintJoinCond:              "AND r.table_name = f.table_name",
		}),
		infos.WithSystemSchemas([]string{"mysql", "information_schema", "performance_schema", "sys"}),
		infos.WithCurrentSchema("COALESCE(DATABASE(), '%')"),
	)
	// NewCompleter for MySQL databases
	NewCompleter = func(db drivers.DB, opts ...completer.Option) readline.AutoCompleter {
		readerOpts := []metadata.ReaderOption{
			// this needs to be relatively low, since autocomplete is very interactive
			metadata.WithTimeout(3 * time.Second),
			metadata.WithLimit(1000),
		}
		reader := NewReader(db, readerOpts...)
		opts = append([]completer.Option{
			completer.WithReader(reader),
			completer.WithDB(db),
			completer.WithSQLStartCommands(append(completer.CommonSqlStartCommands, "USE")),
			completer.WithBeforeComplete(complete(reader)),
		}, opts...)
		return completer.NewDefaultCompleter(opts...)
	}
)

func complete(reader metadata.Reader) completer.CompleteFunc {
	return func(previousWords []string, text []rune) [][]rune {
		if completer.TailMatches(completer.IGNORE_CASE, previousWords, `USE`) {
			return completeWithSchemas(reader, text)
		}
		return nil
	}
}

func completeWithSchemas(reader metadata.Reader, text []rune) [][]rune {
	schemaNames := []string{}
	schemas, err := reader.(metadata.SchemaReader).Schemas(metadata.Filter{WithSystem: true})
	if err != nil {
		return nil
	}
	for schemas.Next() {
		schemaNames = append(schemaNames, schemas.Get().Schema)
	}
	return completer.CompleteFromList(text, schemaNames...)
}
