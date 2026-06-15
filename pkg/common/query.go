package common

// SortColumn represents a single column to sort by
type SortColumn struct {
	Column    string // Column name to sort by
	Direction string // Sort direction: ASC or DESC
}

// QueryOptions provides pagination, sorting, and search parameters for query operations
// Supports both offset-based pagination (for SQL databases) and cursor-based pagination (for Cassandra, etc.)
type QueryOptions struct {
	// Offset-based pagination (for SQL databases where total count is known)
	Offset int // Pagination offset (ignored if Cursor is provided)
	Limit  int // Maximum number of records to return

	// Cursor-based pagination (for databases like Cassandra where total count is expensive/unknown)
	Cursor string // Opaque cursor token from previous response (takes precedence over Offset)

	// Result cache key for pre-computed result sets
	// When provided, the store may use a cached/prepared result set for faster subsequent queries
	ResultCacheKey string // Optional cache key from previous response

	// Sorting (supports multiple columns)
	// Default: [{Column: "created_at", Direction: "DESC"}, {Column: "iid", Direction: "ASC"}]
	SortBy []SortColumn

	// Full-text search query across string and JSONB fields
	Search string

	// Additional filters (key-value pairs for exact match filtering)
	Filters map[string]string
}

// ListingAllowlist declares a backing table's listing contract:
//   - SortableFields: map of client-facing field name (proto/REST) -> SQL column. Anything not in the map is rejected.
//   - SearchableColumns: columns BuildSearchClause runs ILIKE against (covers every visible string + JSONB column).
//   - Default: the SortColumn slice applied when the request carries zero sort fields.
//
// Each store declares one of these once and the gRPC + REST handlers consult it
// to validate inbound sort fields and to fall back to a sane default.
type ListingAllowlist struct {
	SortableFields    map[string]string
	SearchableColumns []string
	Default           []SortColumn
}

// QueryResponse represents a paginated query response
type QueryResponse struct {
	// Total count of records matching the query (may be -1 for cursor-based pagination where count is unknown)
	TotalCount int

	// Next cursor token for cursor-based pagination
	// Client should pass this as Cursor in the next QueryOptions to get the next page
	NextCursor string

	// Result cache key for this query
	// The store may cache intermediate results and return this key
	// Client can pass this as ResultCacheKey in subsequent requests for faster queries
	ResultCacheKey string

	// Whether there are more results available
	HasMore bool
}
