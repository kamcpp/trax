package common

import (
	"fmt"
	"strings"
)

// BuildOrderByClause builds SQL ORDER BY clause from QueryOptions
// tableAlias is optional - if provided, columns will be qualified with it
func BuildOrderByClause(options *QueryOptions, tableAlias string) string {
	prefix := ""
	if tableAlias != "" {
		prefix = tableAlias + "."
	}

	if options == nil || len(options.SortBy) == 0 {
		// Default sort: created_at DESC, iid ASC
		return fmt.Sprintf("ORDER BY %screated_at DESC, %siid ASC", prefix, prefix)
	}

	var sortClauses []string
	for _, sort := range options.SortBy {
		direction := strings.ToUpper(sort.Direction)
		if direction != "ASC" && direction != "DESC" {
			direction = "ASC"
		}
		sortClauses = append(sortClauses, fmt.Sprintf("%s%s %s", prefix, sort.Column, direction))
	}

	return "ORDER BY " + strings.Join(sortClauses, ", ")
}

// BuildPaginationClause builds SQL LIMIT/OFFSET clause from QueryOptions
func BuildPaginationClause(options *QueryOptions) string {
	if options == nil {
		return ""
	}

	if options.Limit <= 0 {
		return ""
	}

	clause := fmt.Sprintf("LIMIT %d", options.Limit)
	if options.Offset > 0 {
		clause += fmt.Sprintf(" OFFSET %d", options.Offset)
	}

	return clause
}

// BuildFilterClause builds a SQL WHERE clause for exact-match filters from QueryOptions.Filters.
// Only columns present in allowedColumns are applied (to prevent SQL injection).
// Returns the clause (e.g. "WHERE col1 = $1 AND col2 = $2") and the positional arguments.
// startArgIndex specifies the first $N placeholder index (use 1 if no prior args, or len(priorArgs)+1).
func BuildFilterClause(options *QueryOptions, allowedColumns map[string]bool, startArgIndex int) (string, []interface{}) {
	if options == nil || len(options.Filters) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}
	argIndex := startArgIndex

	for col, val := range options.Filters {
		if !allowedColumns[col] {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = $%d", col, argIndex))
		args = append(args, val)
		argIndex++
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// BuildSearchClause builds SQL WHERE clause for full-text search
// Returns the WHERE clause and the arguments to be used with the query
func BuildSearchClause(options *QueryOptions, searchColumns []string) (string, []interface{}) {
	if options == nil || options.Search == "" || len(searchColumns) == 0 {
		return "", nil
	}

	searchPattern := "%" + options.Search + "%"
	var conditions []string
	var args []interface{}

	// Common JSONB column names that need casting
	jsonbColumns := map[string]bool{
		"display_names":      true,
		"descriptions":       true,
		"labels":             true,
		"tags":               true,
		"metadata":           true,
		"identifiers":        true,
		"classes":            true,
		"deployment_details": true,
	}

	argIndex := 1
	for _, column := range searchColumns {
		// Extract the base column name (without table alias)
		baseColumn := column
		if dotIndex := strings.LastIndex(column, "."); dotIndex != -1 {
			baseColumn = column[dotIndex+1:]
		}

		// Handle JSONB columns (either containing -> or being a known JSONB column)
		if strings.Contains(column, "->") || jsonbColumns[baseColumn] {
			conditions = append(conditions, fmt.Sprintf("CAST(%s AS TEXT) ILIKE $%d", column, argIndex))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", column, argIndex))
		}
		args = append(args, searchPattern)
		argIndex++
	}

	whereClause := "WHERE (" + strings.Join(conditions, " OR ") + ")"
	return whereClause, args
}
