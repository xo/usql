# D3. No metadata reader is written in usql

Status: Decided.

Ken decided on 2026-09-27 that all database metadata moves into `dbmeta`, so
no metadata reader is written in usql. The question came up for the SurrealDB
driver, which has none.

This extends the freeze that W21 records. On 2026-09-26 Ken decided that
`drivers/metadata` takes no more changes before the migration to `dbmeta`.
That covered the existing readers. This decision covers new ones too.

## What it means for a driver

- A new driver sets no `NewMetadataReader` and no `NewMetadataWriter`. Its
  describe commands, such as `\d`, `\dt`, `\l` and `\dp`, report that they are
  not supported until `dbmeta` answers them.
- A gap in an existing reader, such as a missing `CatalogReader` or
  `PrivilegeSummaryReader`, is fixed in `dbmeta`, not here. W18 measured those
  gaps, and it is superseded by W21, which holds the migration.
- Step 5 of `docs/DRIVER.md` no longer asks for a reader. It says where the
  metadata for a new database goes instead.

The SurrealDB driver is the first driver added under this rule.
