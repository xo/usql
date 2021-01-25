-- echo all
\set ECHO all
-- conditional variables display FALSE when name is not set
\unset foo
\echo :{?foo}
-- conditional variables display TRUE when name is set
\set foo 'bar'
\echo :{?foo}
-- single quoted strings will decode '' as ' and decode \n, \t, \b, \r, \f, \digits octals, \xdigits (standard escapes)
\set foo 'bar''bar\r\n'
-- single quoted variables escape ' but does not escape special characters
\echo :'foo'
-- double quoted variables do not escape ' or special characters
\echo :"foo"
-- single quoted strings decode any other standard escape (\<char>) as literal
\set foo 'bar\'''bar'
\echo :foo
\echo :'foo'
-- single quoted variables escape \ with E'' style strings
\set foo 'bar\\\''
\echo :foo
\echo :'foo'
\echo :"foo"
-- backticks interpolate unquoted variables
\set foo 'bar'
\echo `echo :foo`
-- backticks interpolate single quoted variables
\echo `echo :'foo'`
-- backticks do not interpolate double quoted variables
\echo `echo :"foo"`
-- backticks have error messages for single quoted variables containing \r or \n when using :'' syntax
\set foo 'bar\r\n'
\echo `echo :'foo'`
