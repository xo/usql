# Use a pure Go Dameng driver

Dameng support uses the MIT-licensed `database/sql` driver from `github.com/godoes/gorm-dameng/dm8` instead of DBX's Java/JDBC agent. The initial target is DM8 and the driver belongs to usql's `most` build group. This keeps usql a single binary and gives Dameng the same commands, permissions, metadata operations, output formats, and named-connection workflow as MySQL.

The `dm`, `dm8`, and `dameng` schemes and their DSN generation come from `dburl` v0.25.0, so usql registers only the driver behavior. `sslmode` is not translated: DM8's native `sslFilesPath`, `sslCertPath`, and `sslKeyPath` options are passed through unchanged.
