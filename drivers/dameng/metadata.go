package dameng

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/xo/usql/drivers"
	"github.com/xo/usql/drivers/metadata"
	orameta "github.com/xo/usql/drivers/metadata/oracle"
)

const (
	// systemSchemas matches the Oracle-compatible schemas hidden by the reused base reader.
	systemSchemas = "'CTXSYS', 'FLOWS_FILES', 'MDSYS', 'OUTLN', 'SYS', 'SYSTEM', 'XDB', 'XS$NULL'"
	// systemMaterializedViewsQuery identifies DM8 materialized views from its native object flags.
	systemMaterializedViewsQuery = `SELECT
  schema_object.NAME AS owner,
  materialized_view.NAME AS mview_name
FROM SYS.SYSOBJECTS materialized_view
JOIN SYS.SYSOBJECTS schema_object
  ON schema_object.ID = materialized_view.SCHID
 AND schema_object.TYPE$ = 'SCH'
WHERE materialized_view.TYPE$ = 'SCHOBJ'
  AND materialized_view.SUBTYPE$ = 'VIEW'
  AND (materialized_view.INFO1 & 0x200) > 0`
	// accessibleMaterializedViewsQuery is the restricted-account fallback used when SYSOBJECTS is denied.
	accessibleMaterializedViewsQuery = `SELECT DISTINCT
  OWNER,
  NAME AS mview_name
FROM ALL_DEPENDENCIES
WHERE TYPE IN ('MATERIALIZED VIEW', 'MATERIALIZED_VIEW')`
)

// objectKey uniquely identifies an Oracle-compatible schema object.
type objectKey struct {
	// schema is the owning DM8 schema name.
	schema string
	// name is the object name inside schema.
	name string
}

// dmMetadataReader overrides only DM8-specific metadata and delegates all other capabilities to Oracle.
type dmMetadataReader struct {
	// LoggingReader applies usql's shared query logging, dry-run, and timeout behavior to DM catalogs.
	metadata.LoggingReader
	// tables is the existing Oracle-compatible table reader reused for filtering and synonyms.
	tables metadata.TableReader
}

var (
	// Compile-time checks keep the two DM-specific metadata capabilities registered with PluginReader.
	_ metadata.TableReader            = &dmMetadataReader{}
	_ metadata.PrivilegeSummaryReader = &dmMetadataReader{}
)

// newMetadataReader composes Oracle-compatible metadata with DM8 table and privilege overrides.
func newMetadataReader(db drivers.DB, opts ...metadata.ReaderOption) metadata.Reader {
	// oracleReader supplies every metadata operation that DM8 already implements compatibly.
	oracleReader := orameta.NewReader()(db, opts...)
	// dmReader owns only behavior that differs from Oracle or is missing there.
	dmReader := &dmMetadataReader{
		LoggingReader: metadata.NewLoggingReader(db, opts...),
		tables:        oracleReader.(metadata.TableReader),
	}
	// Later readers override matching capabilities while all untouched operations remain Oracle-backed.
	return metadata.NewPluginReader(oracleReader, dmReader)
}

// Tables classifies DM8 materialized views while preserving the Oracle reader's filters and synonyms.
func (r dmMetadataReader) Tables(filter metadata.Filter) (*metadata.TableSet, error) {
	// Skip DM-specific work when the caller asks only for unrelated object types.
	if !needsMaterializedViewClassification(filter.Types) {
		return r.tables.Tables(filter)
	}

	// baseFilter broadens MATERIALIZED VIEW to VIEW because DM8 may expose both as VIEW.
	baseFilter := filter
	baseFilter.Types = materializedViewCompatibleTypes(filter.Types)
	// baseTables reuses the existing Oracle query and all of its filtering behavior.
	baseTables, err := r.tables.Tables(baseFilter)
	// Return base catalog failures before attempting a secondary classification query.
	if err != nil {
		return nil, err
	}
	// Close releases the base metadata result after it has been copied and classified.
	defer baseTables.Close()

	// materializedViews contains native DM8 object identities that need type correction.
	materializedViews, err := r.materializedViews(filter)
	// A non-permission catalog error indicates a real metadata failure and must remain visible.
	if err != nil {
		return nil, err
	}
	// allowedTypes preserves the caller's original type filter after the compatibility broadening.
	allowedTypes := make(map[string]struct{}, len(filter.Types))
	// Record every requested type for constant-time post-classification filtering.
	for _, objectType := range filter.Types {
		allowedTypes[strings.ToUpper(objectType)] = struct{}{}
	}
	// results stores copied table rows because the base result set is closed on return.
	results := []metadata.Table{}
	// Iterate over every Oracle-compatible row to correct DM8's VIEW representation.
	for baseTables.Next() {
		// table is a value copy that can be safely reclassified without mutating the base set.
		table := *baseTables.Get()
		// Only VIEW is ambiguous; native MATERIALIZED VIEW and all other types remain unchanged.
		if table.Type == "VIEW" {
			// key matches the catalog identity returned by DM8's materialized-view query.
			key := objectKey{schema: table.Schema, name: table.Name}
			// A matching native flag proves the compatible VIEW is materialized.
			if _, ok := materializedViews[key]; ok {
				table.Type = "MATERIALIZED VIEW"
			}
		}
		// A caller without a type filter accepts every classified object.
		if len(allowedTypes) == 0 {
			results = append(results, table)
			continue
		}
		// Skip objects whose corrected type no longer matches the original request.
		if _, ok := allowedTypes[strings.ToUpper(table.Type)]; !ok {
			continue
		}
		results = append(results, table)
	}
	// Propagate deferred row iteration failures instead of returning a partial object list.
	if err := baseTables.Err(); err != nil {
		return nil, err
	}
	return metadata.NewTableSet(results), nil
}

// PrivilegeSummaries combines DM8 object and column grants using the shared metadata model.
func (r dmMetadataReader) PrivilegeSummaries(filter metadata.Filter) (*metadata.PrivilegeSummarySet, error) {
	// tables provides the complete filtered object list, including objects without explicit grants.
	tables, err := r.Tables(filter)
	// Table discovery errors prevent a trustworthy privilege summary.
	if err != nil {
		return nil, err
	}
	// Close releases the table result after summary seeds are copied.
	defer tables.Close()

	// summaries stores one shared-model privilege record per database object.
	summaries := []metadata.PrivilegeSummary{}
	// indices maps privilege catalog rows back to their owning summary.
	indices := map[objectKey]int{}
	// Seed summaries from tables so objects with no explicit privileges remain visible to \dp.
	for tables.Next() {
		// table is the current filtered DM8 object.
		table := tables.Get()
		// summary initializes the shared privilege collections expected by DefaultWriter.
		summary := metadata.PrivilegeSummary{
			Catalog:          table.Catalog,
			Schema:           table.Schema,
			Name:             table.Name,
			ObjectType:       table.Type,
			ObjectPrivileges: metadata.ObjectPrivileges{},
			ColumnPrivileges: metadata.ColumnPrivileges{},
		}
		summaries = append(summaries, summary)
		indices[objectKey{schema: table.Schema, name: table.Name}] = len(summaries) - 1
	}
	// Propagate table row iteration failures before querying grant catalogs.
	if err := tables.Err(); err != nil {
		return nil, err
	}
	// An empty object set needs no privilege catalog scans.
	if len(summaries) == 0 {
		return metadata.NewPrivilegeSummarySet(summaries), nil
	}
	// Append DM8 object-level grants to their seeded summaries.
	if err := r.appendObjectPrivileges(summaries, indices, filter); err != nil {
		return nil, err
	}
	// Append DM8 column-level grants to the same shared summaries.
	if err := r.appendColumnPrivileges(summaries, indices, filter); err != nil {
		return nil, err
	}
	return metadata.NewPrivilegeSummarySet(summaries), nil
}

// materializedViews returns native DM8 object identities, with a restricted-catalog fallback.
func (r dmMetadataReader) materializedViews(filter metadata.Filter) (map[objectKey]struct{}, error) {
	// objects first uses SYSOBJECTS because it exposes DM8's authoritative materialized-view flag.
	objects, err := r.queryMaterializedViews(systemMaterializedViewsQuery, "schema_object.NAME", "materialized_view.NAME", filter)
	// A successful system query provides the complete classification result.
	if err == nil {
		return objects, nil
	}
	// Non-permission failures must not be hidden behind a less authoritative query.
	if !isMetadataPermissionError(err) {
		return nil, err
	}
	// accessibleObjects retries through the catalog available to ordinary DM8 accounts.
	accessibleObjects, accessibleErr := r.queryMaterializedViews(accessibleMaterializedViewsQuery, "OWNER", "NAME", filter)
	// A successful fallback preserves classification without requiring elevated SYS access.
	if accessibleErr == nil {
		return accessibleObjects, nil
	}
	// Non-permission fallback failures remain actionable metadata errors.
	if !isMetadataPermissionError(accessibleErr) {
		return nil, accessibleErr
	}
	// ponytail: restricted accounts return unclassified base metadata; add USER_MVIEWS only if this occurs in production.
	return map[objectKey]struct{}{}, nil
}

// queryMaterializedViews executes one DM8 catalog strategy and returns its object identities.
func (r dmMetadataReader) queryMaterializedViews(query, schemaColumn, nameColumn string, filter metadata.Filter) (map[objectKey]struct{}, error) {
	// conditions and values apply the same schema, visibility, and name scope as the base table query.
	conditions, values := metadataConditions(filter, schemaColumn, nameColumn)
	// Append additional filters to the catalog query's existing WHERE clause.
	if len(conditions) != 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY owner, mview_name"
	// rows executes through LoggingReader so dry-run, timeout, and query logging remain consistent.
	rows, closeRows, err := r.Query(query, values...)
	// Dry-run and empty-query behavior represent an empty classification set.
	if err == sql.ErrNoRows {
		return map[objectKey]struct{}{}, nil
	}
	// Return catalog errors so the caller can decide whether a permission fallback is safe.
	if err != nil {
		return nil, err
	}
	// Close releases the database rows once every object identity has been scanned.
	defer closeRows()

	// objects stores each materialized-view identity for constant-time table classification.
	objects := map[objectKey]struct{}{}
	// Iterate over every matching native materialized-view row.
	for rows.Next() {
		// schema and name receive the two projected object identity columns.
		var schema, name string
		// A malformed catalog row prevents safe object classification.
		if err := rows.Scan(&schema, &name); err != nil {
			return nil, err
		}
		objects[objectKey{schema: schema, name: name}] = struct{}{}
	}
	// Return deferred driver errors instead of a partial identity map.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return objects, nil
}

// appendObjectPrivileges attaches rows from DM8's ALL_TAB_PRIVS catalog.
func (r dmMetadataReader) appendObjectPrivileges(summaries []metadata.PrivilegeSummary, indices map[objectKey]int, filter metadata.Filter) error {
	// query uses OWNER because that is DM8's object-privilege schema column.
	query := `SELECT OWNER, TABLE_NAME, GRANTEE, GRANTOR, PRIVILEGE, GRANTABLE
FROM ALL_TAB_PRIVS`
	// conditions and values restrict the catalog scan to the requested object scope.
	conditions, values := metadataConditions(filter, "OWNER", "TABLE_NAME")
	// Add a WHERE clause only when the caller supplied or implied catalog filters.
	if len(conditions) != 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY OWNER, TABLE_NAME, GRANTEE, GRANTOR, PRIVILEGE"
	// rows reads the DM8 object-level grant catalog through shared reader options.
	rows, closeRows, err := r.Query(query, values...)
	// A catalog with no privilege rows leaves the seeded summaries unchanged.
	if err == sql.ErrNoRows {
		return nil
	}
	// Return query failures rather than presenting incomplete privileges.
	if err != nil {
		return err
	}
	// Close releases object privilege rows after aggregation.
	defer closeRows()

	// Iterate over every explicit object-level grant.
	for rows.Next() {
		// schema, name, grantee, grantor, privilege, and grantable receive one DM8 grant row.
		var schema, name, grantee, grantor, privilege, grantable string
		// A malformed privilege row prevents a trustworthy summary.
		if err := rows.Scan(&schema, &name, &grantee, &grantor, &privilege, &grantable); err != nil {
			return err
		}
		// index locates the pre-filtered object summary for this grant.
		index, ok := indices[objectKey{schema: schema, name: name}]
		// Ignore grants for object types excluded by the table metadata filter.
		if !ok {
			continue
		}
		// objectPrivilege converts DM8's row into usql's shared privilege representation.
		objectPrivilege := metadata.ObjectPrivilege{
			Grantee:       grantee,
			Grantor:       grantor,
			PrivilegeType: privilege,
			IsGrantable:   strings.EqualFold(grantable, "YES"),
		}
		summaries[index].ObjectPrivileges = append(summaries[index].ObjectPrivileges, objectPrivilege)
	}
	return rows.Err()
}

// appendColumnPrivileges attaches rows from DM8's ALL_COL_PRIVS catalog.
func (r dmMetadataReader) appendColumnPrivileges(summaries []metadata.PrivilegeSummary, indices map[objectKey]int, filter metadata.Filter) error {
	// query uses TABLE_SCHEMA because DM8 names the column-privilege schema differently from ALL_TAB_PRIVS.
	query := `SELECT TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, GRANTEE, GRANTOR, PRIVILEGE, GRANTABLE
FROM ALL_COL_PRIVS`
	// conditions and values restrict the catalog scan to the requested object scope.
	conditions, values := metadataConditions(filter, "TABLE_SCHEMA", "TABLE_NAME")
	// Add a WHERE clause only when the caller supplied or implied catalog filters.
	if len(conditions) != 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY TABLE_SCHEMA, TABLE_NAME, COLUMN_NAME, GRANTEE, GRANTOR, PRIVILEGE"
	// rows reads the DM8 column-level grant catalog through shared reader options.
	rows, closeRows, err := r.Query(query, values...)
	// A catalog with no column grants leaves the seeded summaries unchanged.
	if err == sql.ErrNoRows {
		return nil
	}
	// Return query failures rather than presenting incomplete privileges.
	if err != nil {
		return err
	}
	// Close releases column privilege rows after aggregation.
	defer closeRows()

	// Iterate over every explicit column-level grant.
	for rows.Next() {
		// schema, name, column, grantee, grantor, privilege, and grantable receive one DM8 grant row.
		var schema, name, column, grantee, grantor, privilege, grantable string
		// A malformed privilege row prevents a trustworthy summary.
		if err := rows.Scan(&schema, &name, &column, &grantee, &grantor, &privilege, &grantable); err != nil {
			return err
		}
		// index locates the pre-filtered object summary for this grant.
		index, ok := indices[objectKey{schema: schema, name: name}]
		// Ignore grants for object types excluded by the table metadata filter.
		if !ok {
			continue
		}
		// columnPrivilege converts DM8's row into usql's shared privilege representation.
		columnPrivilege := metadata.ColumnPrivilege{
			Column:        column,
			Grantee:       grantee,
			Grantor:       grantor,
			PrivilegeType: privilege,
			IsGrantable:   strings.EqualFold(grantable, "YES"),
		}
		summaries[index].ColumnPrivileges = append(summaries[index].ColumnPrivileges, columnPrivilege)
	}
	return rows.Err()
}

// metadataConditions builds DM8 positional catalog filters shared by classification and privileges.
func metadataConditions(filter metadata.Filter, schemaColumn, nameColumn string) ([]string, []interface{}) {
	// conditions contains SQL predicates in positional-argument order.
	conditions := []string{}
	// values contains the uppercase LIKE patterns bound to each positional placeholder.
	values := []interface{}{}
	// Add the requested schema pattern when the caller supplied one.
	if filter.Schema != "" {
		values = append(values, strings.ToUpper(filter.Schema))
		conditions = append(conditions, fmt.Sprintf("%s LIKE :%d", schemaColumn, len(values)))
	}
	// Hide the same system schemas as the reused Oracle metadata reader.
	if !filter.WithSystem {
		conditions = append(conditions, fmt.Sprintf("%s NOT IN (%s)", schemaColumn, systemSchemas))
	}
	// Restrict results to the connected user when the shared metadata filter requests visibility.
	if filter.OnlyVisible {
		conditions = append(conditions, schemaColumn+" = USER")
	}
	// Add the requested object-name pattern when the caller supplied one.
	if filter.Name != "" {
		values = append(values, strings.ToUpper(filter.Name))
		conditions = append(conditions, fmt.Sprintf("%s LIKE :%d", nameColumn, len(values)))
	}
	return conditions, values
}

// needsMaterializedViewClassification reports whether DM8 VIEW rows can affect the requested result.
func needsMaterializedViewClassification(types []string) bool {
	// An empty type list means the caller accepts all object types, including both view categories.
	if len(types) == 0 {
		return true
	}
	// Inspect each requested type for one of the two categories sharing DM8's VIEW representation.
	for _, objectType := range types {
		// Either view category requires native materialized-view classification.
		if strings.EqualFold(objectType, "VIEW") || strings.EqualFold(objectType, "MATERIALIZED VIEW") {
			return true
		}
	}
	return false
}

// materializedViewCompatibleTypes broadens a type filter for DM8's VIEW compatibility representation.
func materializedViewCompatibleTypes(types []string) []string {
	// A nil or empty type list already permits every object type.
	if len(types) == 0 {
		return types
	}
	// compatibleTypes preserves requested order while preventing duplicate catalog predicates.
	compatibleTypes := make([]string, 0, len(types)+1)
	// seen records normalized object types already added to the broadened filter.
	seen := map[string]struct{}{}
	// Copy each requested type before adding the compatible VIEW form when needed.
	for _, objectType := range types {
		// normalizedType provides case-insensitive duplicate detection.
		normalizedType := strings.ToUpper(objectType)
		// Skip duplicate caller values to keep placeholders and arguments minimal.
		if _, ok := seen[normalizedType]; ok {
			continue
		}
		seen[normalizedType] = struct{}{}
		compatibleTypes = append(compatibleTypes, objectType)
	}
	// A MATERIALIZED VIEW request must also fetch ambiguous VIEW rows for later classification.
	if _, wantsMaterialized := seen["MATERIALIZED VIEW"]; wantsMaterialized {
		// Avoid adding VIEW twice when the caller already requested both categories.
		if _, hasView := seen["VIEW"]; !hasView {
			compatibleTypes = append(compatibleTypes, "VIEW")
		}
	}
	return compatibleTypes
}

// isMetadataPermissionError limits catalog fallback to explicit DM8 permission failures.
func isMetadataPermissionError(err error) bool {
	// message normalizes localized and English DM8 errors for catalog and permission matching.
	message := strings.ToLower(err.Error())
	// A permission token is required so syntax, network, and driver errors are never hidden.
	permissionDenied := strings.Contains(message, "权限") ||
		strings.Contains(message, "privilege") ||
		strings.Contains(message, "permission denied") ||
		strings.Contains(message, "access denied") ||
		strings.Contains(message, "not authorized")
	// Guard against unrelated permission failures before checking metadata object names.
	if !permissionDenied {
		return false
	}
	return strings.Contains(message, "sysobjects") || strings.Contains(message, "all_dependencies")
}
