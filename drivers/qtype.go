package drivers

import (
	"strings"
)

// queryMap is the map of SQL prefixes use as queries.
var queryMap = map[string]bool{
	"WITH":   true,
	"PRAGMA": true,

	"EXPLAIN":  true, // show the execution plan of a statement
	"DESCRIBE": true, // describe (mysql)
	"DESC":     true, // describe (mysql)
	"FETCH":    true, // retrieve rows from a query using a cursor
	"SELECT":   true, // retrieve rows from a table or view
	"SHOW":     true, // show the value of a run-time parameter
	"VALUES":   true, // compute a set of rows
	"LIST":     true, //  list permissions, roles, users [cassandra]

	"EXEC": true, // execute a stored procedure that returns rows (not postgres)
}

// execMap is the map of SQL prefixes to execute.
//
// Unless noted, these are extracted from the PostgreSQL docs.
//
// Note: originally extracted via a script, but maintained by hand as the
// documentation for any new queries introduced by PostgreSQL need to be
// manually scrutinized for variations.
var execMap = map[string]bool{
	// cassandra
	"ALTER KEYSPACE":  true, // alter a keyspace
	"CREATE KEYSPACE": true, // create a keyspace
	"DROP KEYSPACE":   true, // drop a keyspace
	"BEGIN BATCH":     true, // begin batch
	"APPLY BATCH":     true, // apply batch

	// ql
	"BEGIN TRANSACTION": true, // begin batch

	// postgresql
	"ABORT":                            true, // abort the current transaction
	"ALTER AGGREGATE":                  true, // change the definition of an aggregate function
	"ALTER COLLATION":                  true, // change the definition of a collation
	"ALTER CONVERSION":                 true, // change the definition of a conversion
	"ALTER DATABASE":                   true, // change a database
	"ALTER DEFAULT PRIVILEGES":         true, // define default access privileges
	"ALTER DOMAIN":                     true, // change the definition of a domain
	"ALTER EVENT TRIGGER":              true, // change the definition of an event trigger
	"ALTER EXTENSION":                  true, // change the definition of an extension
	"ALTER FOREIGN DATA WRAPPER":       true, // change the definition of a foreign-data wrapper
	"ALTER FOREIGN TABLE":              true, // change the definition of a foreign table
	"ALTER FUNCTION":                   true, // change the definition of a function
	"ALTER GROUP":                      true, // change role name or membership
	"ALTER INDEX":                      true, // change the definition of an index
	"ALTER LANGUAGE":                   true, // change the definition of a procedural language
	"ALTER LARGE OBJECT":               true, // change the definition of a large object
	"ALTER MATERIALIZED VIEW":          true, // change the definition of a materialized view
	"ALTER OPERATOR CLASS":             true, // change the definition of an operator class
	"ALTER OPERATOR FAMILY":            true, // change the definition of an operator family
	"ALTER OPERATOR":                   true, // change the definition of an operator
	"ALTER POLICY":                     true, // change the definition of a row level security policy
	"ALTER ROLE":                       true, // change a database role
	"ALTER RULE":                       true, // change the definition of a rule
	"ALTER SCHEMA":                     true, // change the definition of a schema
	"ALTER SEQUENCE":                   true, // change the definition of a sequence generator
	"ALTER SERVER":                     true, // change the definition of a foreign server
	"ALTER SYSTEM":                     true, // change a server configuration parameter
	"ALTER TABLESPACE":                 true, // change the definition of a tablespace
	"ALTER TABLE":                      true, // change the definition of a table
	"ALTER TEXT SEARCH CONFIGURATION":  true, // change the definition of a text search configuration
	"ALTER TEXT SEARCH DICTIONARY":     true, // change the definition of a text search dictionary
	"ALTER TEXT SEARCH PARSER":         true, // change the definition of a text search parser
	"ALTER TEXT SEARCH TEMPLATE":       true, // change the definition of a text search template
	"ALTER TRIGGER":                    true, // change the definition of a trigger
	"ALTER TYPE":                       true, // change the definition of a type
	"ALTER USER MAPPING":               true, // change the definition of a user mapping
	"ALTER USER":                       true, // change a database role
	"ALTER VIEW":                       true, // change the definition of a view
	"ANALYZE":                          true, // collect statistics about a database
	"BEGIN":                            true, // start a transaction block
	"CHECKPOINT":                       true, // force a transaction log checkpoint
	"CLOSE":                            true, // close a cursor
	"CLUSTER":                          true, // cluster a table according to an index
	"COMMENT":                          true, // define or change the comment of an object
	"COMMIT PREPARED":                  true, // commit a transaction that was earlier prepared for two-phase commit
	"COMMIT":                           true, // commit the current transaction
	"COPY":                             true, // copy data between a file and a table
	"CREATE ACCESS METHOD":             true, // define a new access method
	"CREATE AGGREGATE":                 true, // define a new aggregate function
	"CREATE CAST":                      true, // define a new cast
	"CREATE COLLATION":                 true, // define a new collation
	"CREATE CONVERSION":                true, // define a new encoding conversion
	"CREATE DATABASE":                  true, // create a new database
	"CREATE DOMAIN":                    true, // define a new domain
	"CREATE EVENT TRIGGER":             true, // define a new event trigger
	"CREATE EXTENSION":                 true, // install an extension
	"CREATE FOREIGN DATA WRAPPER":      true, // define a new foreign-data wrapper
	"CREATE FOREIGN TABLE":             true, // define a new foreign table
	"CREATE FUNCTION":                  true, // define a new function
	"CREATE GROUP":                     true, // define a new database role
	"CREATE INDEX":                     true, // define a new index
	"CREATE LANGUAGE":                  true, // define a new procedural language
	"CREATE MATERIALIZED VIEW":         true, // define a new materialized view
	"CREATE OPERATOR CLASS":            true, // define a new operator class
	"CREATE OPERATOR FAMILY":           true, // define a new operator family
	"CREATE OPERATOR":                  true, // define a new operator
	"CREATE POLICY":                    true, // define a new row level security policy for a table
	"CREATE ROLE":                      true, // define a new database role
	"CREATE RULE":                      true, // define a new rewrite rule
	"CREATE SCHEMA":                    true, // define a new schema
	"CREATE SEQUENCE":                  true, // define a new sequence generator
	"CREATE SERVER":                    true, // define a new foreign server
	"CREATE STATISTICS":                true, // define extended statistics
	"CREATE SUBSCRIPTION":              true, // define a new subscription
	"CREATE TABLE AS":                  true, // define a new table from the results of a query
	"CREATE TABLESPACE":                true, // define a new tablespace
	"CREATE TABLE":                     true, // define a new table
	"CREATE TEXT SEARCH CONFIGURATION": true, // define a new text search configuration
	"CREATE TEXT SEARCH DICTIONARY":    true, // define a new text search dictionary
	"CREATE TEXT SEARCH PARSER":        true, // define a new text search parser
	"CREATE TEXT SEARCH TEMPLATE":      true, // define a new text search template
	"CREATE TRANSFORM":                 true, // define a new transform
	"CREATE TRIGGER":                   true, // define a new trigger
	"CREATE TYPE":                      true, // define a new data type
	"CREATE USER MAPPING":              true, // define a new mapping of a user to a foreign server
	"CREATE USER":                      true, // define a new database role
	"CREATE VIEW":                      true, // define a new view
	"DEALLOCATE":                       true, // deallocate a prepared statement
	"DECLARE":                          true, // define a cursor
	"DELETE":                           true, // delete rows of a table
	"DISCARD":                          true, // discard session state
	"DO":                               true, // execute an anonymous code block
	"DROP ACCESS METHOD":               true, // remove an access method
	"DROP AGGREGATE":                   true, // remove an aggregate function
	"DROP CAST":                        true, // remove a cast
	"DROP COLLATION":                   true, // remove a collation
	"DROP CONVERSION":                  true, // remove a conversion
	"DROP DATABASE":                    true, // remove a database
	"DROP DOMAIN":                      true, // remove a domain
	"DROP EVENT TRIGGER":               true, // remove an event trigger
	"DROP EXTENSION":                   true, // remove an extension
	"DROP FOREIGN DATA WRAPPER":        true, // remove a foreign-data wrapper
	"DROP FOREIGN TABLE":               true, // remove a foreign table
	"DROP FUNCTION":                    true, // remove a function
	"DROP GROUP":                       true, // remove a database role
	"DROP INDEX":                       true, // remove an index
	"DROP LANGUAGE":                    true, // remove a procedural language
	"DROP MATERIALIZED VIEW":           true, // remove a materialized view
	"DROP OPERATOR CLASS":              true, // remove an operator class
	"DROP OPERATOR FAMILY":             true, // remove an operator family
	"DROP OPERATOR":                    true, // remove an operator
	"DROP OWNED":                       true, // remove database objects owned by a database role
	"DROP POLICY":                      true, // remove a row level security policy from a table
	"DROP PUBLICATION":                 true, // remove a publication
	"DROP ROLE":                        true, // remove a database role
	"DROP RULE":                        true, // remove a rewrite rule
	"DROP SCHEMA":                      true, // remove a schema
	"DROP SEQUENCE":                    true, // remove a sequence
	"DROP SERVER":                      true, // remove a foreign server descriptor
	"DROP STATISTICS":                  true, // remove extended statistics
	"DROP SUBSCRIPTION":                true, // remove a subscription
	"DROP TABLESPACE":                  true, // remove a tablespace
	"DROP TABLE":                       true, // remove a table
	"DROP TEXT SEARCH CONFIGURATION":   true, // remove a text search configuration
	"DROP TEXT SEARCH DICTIONARY":      true, // remove a text search dictionary
	"DROP TEXT SEARCH PARSER":          true, // remove a text search parser
	"DROP TEXT SEARCH TEMPLATE":        true, // remove a text search template
	"DROP TRANSFORM":                   true, // remove a transform
	"DROP TRIGGER":                     true, // remove a trigger
	"DROP TYPE":                        true, // remove a data type
	"DROP USER MAPPING":                true, // remove a user mapping for a foreign server
	"DROP USER":                        true, // remove a database role
	"DROP VIEW":                        true, // remove a view
	"END":                              true, // commit the current transaction
	"EXECUTE":                          true, // execute a prepared statement
	"GRANT":                            true, // define access privileges
	"IMPORT FOREIGN SCHEMA":            true, // import table definitions from a foreign server
	"INSERT":                           true, // create new rows in a table
	"LISTEN":                           true, // listen for a notification
	"LOAD":                             true, // load a shared library file
	"LOCK":                             true, // lock a table
	"MOVE":                             true, // position a cursor
	"NOTIFY":                           true, // generate a notification
	"PREPARE TRANSACTION":              true, // prepare the current transaction for two-phase commit
	"PREPARE":                          true, // prepare a statement for execution
	"REASSIGN OWNED":                   true, // change the ownership of database objects owned by a database role
	"REFRESH MATERIALIZED VIEW":        true, // replace the contents of a materialized view
	"REINDEX":                          true, // rebuild indexes
	"RELEASE":                          true, // destroy a previously defined savepoint
	"RESET":                            true, // restore the value of a run-time parameter to the default value
	"REVOKE":                           true, // remove access privileges
	"ROLLBACK PREPARED":                true, // cancel a transaction that was earlier prepared for two-phase commit
	"ROLLBACK TO SAVEPOINT":            true, // roll back to a savepoint
	"ROLLBACK":                         true, // abort the current transaction
	"SAVEPOINT":                        true, // define a new savepoint within the current transaction
	"SECURITY LABEL":                   true, // define or change a security label applied to an object
	"SELECT INTO":                      true, // define a new table from the results of a query
	"SET CONSTRAINTS":                  true, // set constraint check timing for the current transaction
	"SET ROLE":                         true, // set the current user identifier of the current session
	"SET SESSION AUTHORIZATION":        true, // set the session user identifier and the current user identifier of the current session
	"SET TRANSACTION":                  true, // set the characteristics of the current transaction
	"SET":                              true, // change a run-time parameter
	"START TRANSACTION":                true, // start a transaction block
	"TRUNCATE":                         true, // empty a table or set of tables
	"UNLISTEN":                         true, // stop listening for a notification
	"UPDATE":                           true, // update rows of a table
	"VACUUM":                           true, // garbage-collect and optionally analyze a database
}

// createIgnore are parts of the query exec type after CREATE to ignore.
var createIgnore = map[string]bool{
	"DEFAULT":    true,
	"GLOBAL":     true,
	"LOCAL":      true,
	"OR":         true,
	"PROCEDURAL": true,
	"RECURSIVE":  true,
	"REPLACE":    true,
	"TEMPORARY":  true,
	"TEMP":       true,
	"TRUSTED":    true,
	"UNIQUE":     true,
	"UNLOGGED":   true,
}

// QueryExecType is the default way to determine the "EXEC" prefix for a SQL
// query and whether or not it should be Exec'd or Query'd.
func QueryExecType(prefix, sqlstr string) (string, bool) {
	if prefix == "" {
		return "EXEC", false
	}

	s := strings.Split(prefix, " ")
	if len(s) > 0 {
		// check query map
		if _, ok := queryMap[s[0]]; ok {
			typ := s[0]
			switch {
			case typ == "SELECT" && len(s) >= 2 && s[1] == "INTO":
				return "SELECT INTO", false
			case typ == "PRAGMA":
				return typ, !strings.ContainsRune(sqlstr, '=')
			}
			return typ, true
		}

		// normalize prefixes
		switch s[0] {
		// CREATE statements have a large number of variants
		case "CREATE":
			n := []string{"CREATE"}
			for _, x := range s[1:] {
				if _, ok := createIgnore[x]; ok {
					continue
				}
				n = append(n, x)
			}
			s = n

		case "DROP":
			// "DROP [PROCEDURAL] LANGUAGE" => "DROP LANGUAGE"
			n := []string{"DROP"}
			for _, x := range s[1:] {
				if x == "PROCEDURAL" {
					continue
				}
				n = append(n, x)
			}
			s = n
		}

		// find longest match
		for i := len(s); i > 0; i-- {
			typ := strings.Join(s[:i], " ")
			if _, ok := execMap[typ]; ok {
				return typ, false
			}
		}
	}

	return s[0], false
}
