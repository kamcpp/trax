package common

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ExecuteSQLFile reads and executes a SQL file, filtering out PostgreSQL-specific commands
func ExecuteSQLFile(t *testing.T, db *sql.DB, filePath string) {
	// Read the SQL file
	content, err := os.ReadFile(filePath)
	require.NoError(t, err, "failed to read SQL file: %s", filePath)

	// Remove PostgreSQL-specific commands that don't work with database/sql
	sqlContent := string(content)

	// Remove \c commands (database connections)
	sqlContent = removePostgreSQLCommands(sqlContent, `\\c\s+\w+;?`)

	// Remove \echo commands
	sqlContent = removePostgreSQLCommands(sqlContent, `\\echo\s+[^;]*;?`)

	// Remove \dn commands
	sqlContent = removePostgreSQLCommands(sqlContent, `\\dn\s*;?`)

	// Remove \dt commands
	sqlContent = removePostgreSQLCommands(sqlContent, `\\dt\s+[^;]*;?`)

	// Remove \l commands
	sqlContent = removePostgreSQLCommands(sqlContent, `\\l\s*;?`)

	// Remove \du commands
	sqlContent = removePostgreSQLCommands(sqlContent, `\\du\s*;?`)

	// Remove \i commands (include file)
	sqlContent = removePostgreSQLCommands(sqlContent, `\\i\s+[^;]*;?`)

	// Remove FK constraints to schemas that may not exist in isolated tests
	// This removes CONSTRAINT + FOREIGN KEY lines that reference accmgr, instrmgr, etc.
	sqlContent = removeCrossSchemaFKConstraints(sqlContent)

	// Split into individual statements, respecting $$ delimited blocks (plpgsql functions)
	statements := splitSQLStatements(sqlContent)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Remove comment-only lines but keep SQL with inline comments
		lines := strings.Split(stmt, "\n")
		var sqlLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Only skip lines that are entirely comments (start with --)
			// Keep lines that have SQL content even if they have inline comments
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				sqlLines = append(sqlLines, line) // Keep original line with indentation
			}
		}

		if len(sqlLines) == 0 {
			continue
		}

		cleanStmt := strings.Join(sqlLines, "\n")
		cleanStmt = strings.TrimSpace(cleanStmt)

		if cleanStmt == "" {
			continue
		}

		// Skip user management and grant statements in tests
		upperStmt := strings.ToUpper(cleanStmt)
		if strings.Contains(upperStmt, "CREATE USER") ||
			strings.Contains(upperStmt, "ALTER USER") ||
			strings.Contains(upperStmt, "GRANT ALL PRIVILEGES") ||
			strings.Contains(upperStmt, "GRANT USAGE ON SCHEMA") ||
			strings.Contains(upperStmt, "GRANT SELECT") ||
			strings.Contains(upperStmt, "SELECT 'CREATE DATABASE") ||
			strings.Contains(upperStmt, "FROM PG_USER") ||
			strings.Contains(upperStmt, "CURRENT_DATABASE()") {
			continue
		}

		_, err := db.Exec(cleanStmt)
		if err != nil {
			t.Logf("Failed statement: %s", cleanStmt)
			require.NoError(t, err, "failed to execute SQL statement")
		}
	}
}

// FindProjectRoot navigates up the directory tree to find the trax project root
func FindProjectRoot(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	// Navigate up to find the trax directory
	projectRoot := wd
	for {
		// Check if we found the project root by looking for key files/directories
		if hasProjectRootMarkers(projectRoot) {
			return projectRoot
		}

		// Also check if path ends with trax (for local development)
		if strings.HasSuffix(projectRoot, filepath.Join("agora", "daemons")) {
			return projectRoot
		}

		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatal("could not find trax project root")
		}
		projectRoot = parent
	}
}

// hasProjectRootMarkers checks if the directory contains key files indicating it's the project root
func hasProjectRootMarkers(dir string) bool {
	// Look for go.mod file with the expected module path
	goModPath := filepath.Join(dir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		if strings.Contains(string(content), "github.com/xshyft/trax") {
			return true
		}
	}

	// Look for the SQL schema file in expected location
	schemaPath := filepath.Join(dir, "deploy", "k8s", "init", "init_trax_pgsql.sql")
	if _, err := os.Stat(schemaPath); err == nil {
		return true
	}

	return false
}

// splitSQLStatements splits SQL content into individual statements, properly handling
// $$ delimited blocks (used for plpgsql function bodies), single-quoted string
// literals (with ” escape), and -- line comments. Semicolons inside any of
// those must NOT be treated as statement terminators.
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inDollarQuote := false
	inLineComment := false
	inSingleQuote := false

	i := 0
	for i < len(content) {
		// Track single-line comments (-- to end of line).
		if !inLineComment && !inDollarQuote && !inSingleQuote &&
			i+1 < len(content) && content[i] == '-' && content[i+1] == '-' {
			inLineComment = true
		}
		if inLineComment && content[i] == '\n' {
			inLineComment = false
		}

		// Single-quoted string literal — postgres uses '' as the escape
		// for a literal apostrophe inside the string, so a doubled quote
		// stays inside the string rather than closing it.
		if !inLineComment && !inDollarQuote && content[i] == '\'' {
			if inSingleQuote && i+1 < len(content) && content[i+1] == '\'' {
				current.WriteByte('\'')
				current.WriteByte('\'')
				i += 2
				continue
			}
			inSingleQuote = !inSingleQuote
			current.WriteByte('\'')
			i++
			continue
		}

		// Check for $$ delimiter
		if !inLineComment && !inSingleQuote &&
			i+1 < len(content) && content[i] == '$' && content[i+1] == '$' {
			current.WriteString("$$")
			inDollarQuote = !inDollarQuote
			i += 2
			continue
		}

		// Split on semicolons only when NOT inside a $$ block, a single-
		// quoted string literal, or a line comment.
		if content[i] == ';' && !inDollarQuote && !inLineComment && !inSingleQuote {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
			i++
			continue
		}

		current.WriteByte(content[i])
		i++
	}

	// Don't forget the last statement (if any)
	if stmt := current.String(); strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}

func removePostgreSQLCommands(content, pattern string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		matched := false

		// Simple pattern matching for PostgreSQL commands
		if strings.Contains(pattern, "\\\\c") && strings.HasPrefix(trimmed, "\\c") {
			matched = true
		} else if strings.Contains(pattern, "\\\\echo") && strings.HasPrefix(trimmed, "\\echo") {
			matched = true
		} else if strings.Contains(pattern, "\\\\dn") && strings.HasPrefix(trimmed, "\\dn") {
			matched = true
		} else if strings.Contains(pattern, "\\\\dt") && strings.HasPrefix(trimmed, "\\dt") {
			matched = true
		} else if strings.Contains(pattern, "\\\\l") && strings.HasPrefix(trimmed, "\\l") {
			matched = true
		} else if strings.Contains(pattern, "\\\\du") && strings.HasPrefix(trimmed, "\\du") {
			matched = true
		} else if strings.Contains(pattern, "\\\\i") && strings.HasPrefix(trimmed, "\\i") {
			matched = true
		}

		if !matched {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// removeCrossSchemaFKConstraints removes FK constraint declarations that reference
// schemas not initialized in isolated tests (accmgr, instrmgr, laser).
// This allows component tests to run without requiring all schemas to be present.
//
// It removes multi-line FK constraints like:
//
//	CONSTRAINT fk_accounts_agora_account
//	FOREIGN KEY (agora_account_iid) REFERENCES accmgr.accounts(iid) ON DELETE SET NULL,
func removeCrossSchemaFKConstraints(content string) string {
	// List of schemas to exclude from FK constraint checks
	// NOTE: Only exclude schemas that are truly not initialized in tests
	// instrmgr and accmgr have been removed as they may be needed in some tests
	excludedSchemas := []string{}

	lines := strings.Split(content, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		upperLine := strings.ToUpper(trimmed)

		// Check if this line starts a FK constraint
		if strings.HasPrefix(upperLine, "CONSTRAINT FK_") {
			// Look ahead to find the FOREIGN KEY line and check if it references an excluded schema
			isExcluded := false
			endIdx := i

			// Search next few lines for FOREIGN KEY ... REFERENCES
			for j := i; j < len(lines) && j < i+5; j++ {
				nextUpper := strings.ToUpper(strings.TrimSpace(lines[j]))
				for _, schema := range excludedSchemas {
					if strings.Contains(nextUpper, "REFERENCES "+schema+".") {
						isExcluded = true
						// Find the end of this FK constraint (line with comma or semicolon)
						for k := j; k < len(lines); k++ {
							endIdx = k
							kTrimmed := strings.TrimSpace(lines[k])
							if strings.HasSuffix(kTrimmed, ",") || strings.HasSuffix(kTrimmed, ";") {
								break
							}
						}
						break
					}
				}
				if isExcluded {
					break
				}
			}

			if isExcluded {
				// Remove trailing comma from the previous line (if any)
				if len(result) > 0 {
					lastIdx := len(result) - 1
					lastLine := result[lastIdx]
					lastTrimmed := strings.TrimSpace(lastLine)
					if strings.HasSuffix(lastTrimmed, ",") {
						// Remove the trailing comma
						result[lastIdx] = strings.TrimRight(lastLine, ",")
					}
				}

				// Skip all lines from i to endIdx (inclusive)
				i = endIdx + 1
				continue
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

// ============================================================================
// PostgreSQL Test Container Management
// ============================================================================

// PostgresTestContainer represents a shared PostgreSQL test container
// that is reused across multiple tests for performance and efficiency.
type PostgresTestContainer struct {
	Container        testcontainers.Container
	Host             string
	Port             string
	User             string
	Password         string
	MasterConnString string // Connection string to 'postgres' database
}

// isRunningInDocker checks if we're running inside a Docker container
func isRunningInDocker() bool {
	// Check for /.dockerenv file (common indicator)
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	// Check for /workspace which is our mounted workspace in the test container
	if _, err := os.Stat("/workspace"); err == nil {
		return true
	}
	return false
}

// StartSharedPostgresContainer starts a single PostgreSQL container for all tests.
// The container uses tmpfs for in-memory storage (fast, no disk I/O).
// This should be called once in TestMain before running tests.
func StartSharedPostgresContainer(ctx context.Context) (*PostgresTestContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "postgres",
		},
		Tmpfs: map[string]string{
			"/var/lib/postgresql/data": "rw", // Use tmpfs for in-memory storage
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2). // PostgreSQL logs this twice during startup
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	var host string
	var port string

	// When running inside Docker (Docker-in-Docker), we need to use the container's
	// internal IP address rather than the host IP, since the mapped port on the host
	// isn't accessible from within another container on the same Docker network.
	if isRunningInDocker() {
		// Get the container's internal IP address
		inspect, err := container.Inspect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect container: %w", err)
		}

		// Get the IP from the bridge network (default Docker network)
		if networkSettings := inspect.NetworkSettings; networkSettings != nil {
			// Try bridge network first
			if bridge, ok := networkSettings.Networks["bridge"]; ok && bridge.IPAddress != "" {
				host = bridge.IPAddress
			} else {
				// Fall back to the first available network
				for _, network := range networkSettings.Networks {
					if network.IPAddress != "" {
						host = network.IPAddress
						break
					}
				}
			}
		}

		if host == "" {
			return nil, fmt.Errorf("failed to get container internal IP address")
		}

		// Use the internal port directly (not the mapped port)
		port = "5432"
	} else {
		// Running on host - use standard host/mapped port
		host, err = container.Host(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get container host: %w", err)
		}

		mappedPort, err := container.MappedPort(ctx, "5432")
		if err != nil {
			return nil, fmt.Errorf("failed to get mapped port: %w", err)
		}
		port = mappedPort.Port()
	}

	// Build connection string for 'postgres' database (used for creating/dropping test databases)
	masterConnString := fmt.Sprintf("postgres://test_user:test_password@%s:%s/postgres?sslmode=disable", host, port)

	return &PostgresTestContainer{
		Container:        container,
		Host:             host,
		Port:             port,
		User:             "test_user",
		Password:         "test_password",
		MasterConnString: masterConnString,
	}, nil
}

// StopSharedPostgresContainer stops and removes the shared PostgreSQL container.
// This should be called once in TestMain after all tests have completed.
func StopSharedPostgresContainer(ctx context.Context, container *PostgresTestContainer) error {
	if container == nil || container.Container == nil {
		return nil
	}

	if err := container.Container.Terminate(ctx); err != nil {
		return fmt.Errorf("failed to terminate PostgreSQL container: %w", err)
	}

	return nil
}

// CreateTestDatabase creates a new isolated database within the shared container.
// Each test should create its own database for complete isolation.
// Returns: (*sql.DB, connectionString, error)
func CreateTestDatabase(ctx context.Context, container *PostgresTestContainer, dbName string) (*sql.DB, string, error) {
	// Connect to 'postgres' database to create new database
	masterDB, err := sql.Open("postgres", container.MasterConnString)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to master database: %w", err)
	}
	defer masterDB.Close()

	// Create the test database
	_, err = masterDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	// Build connection string for the new database
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		container.User, container.Password, container.Host, container.Port, dbName)

	// Connect to the new database
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to database %s: %w", dbName, err)
	}

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, "", fmt.Errorf("failed to ping database %s: %w", dbName, err)
	}

	return db, connString, nil
}

// DropTestDatabase drops a test database from the shared container.
// This should be called in the cleanup function returned by test setup helpers.
func DropTestDatabase(ctx context.Context, container *PostgresTestContainer, dbName string) error {
	// Connect to 'postgres' database
	masterDB, err := sql.Open("postgres", container.MasterConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to master database: %w", err)
	}
	defer masterDB.Close()

	// Terminate existing connections to the database
	terminateQuery := `
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = $1
		  AND pid <> pg_backend_pid()
	`
	_, err = masterDB.ExecContext(ctx, terminateQuery, dbName)
	if err != nil {
		// Non-fatal - database might not have active connections
		// Continue with drop attempt
	}

	// Drop the database
	_, err = masterDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	return nil
}

// GenerateTestDatabaseName creates a unique database name for a test.
// The name includes a timestamp to ensure uniqueness across test runs.
func GenerateTestDatabaseName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// ============================================================================
// Helper Functions
// ============================================================================

// ExecuteSQLFilesInOrder executes multiple SQL files sequentially.
// Useful for initializing schemas that have dependencies (e.g., shared schema first, then component schema).
func ExecuteSQLFilesInOrder(t *testing.T, db *sql.DB, filePaths ...string) error {
	for _, filePath := range filePaths {
		t.Logf("Executing SQL file: %s", filePath)
		ExecuteSQLFile(t, db, filePath)
	}
	return nil
}

// VerifyTableExists checks if a table exists in the specified schema.
// Useful for verifying that schema initialization completed successfully.
func VerifyTableExists(t *testing.T, db *sql.DB, schema, table string) bool {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM pg_catalog.pg_tables
			WHERE schemaname = $1 AND tablename = $2
		)
	`
	var exists bool
	err := db.QueryRow(query, schema, table).Scan(&exists)
	if err != nil {
		t.Logf("Error checking table existence: %v", err)
		return false
	}
	return exists
}

// CountRows returns the number of rows in the specified table.
// Useful for validating test data in assertions.
func CountRows(t *testing.T, db *sql.DB, schema, table string) int {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schema, table)
	var count int
	err := db.QueryRow(query).Scan(&count)
	require.NoError(t, err, "failed to count rows in %s.%s", schema, table)
	return count
}

// ============================================================================
// Schema Initialization Functions
// ============================================================================

// InitializeSharedSchema executes init_trax_pgsql.sql to create the shared.entities table.
// This must be called before initializing any component-specific schema.
func InitializeSharedSchema(t *testing.T, db *sql.DB) error {
	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_trax_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify shared.entities table was created
	if !VerifyTableExists(t, db, "shared", "entities") {
		return fmt.Errorf("shared.entities table was not created")
	}

	return nil
}

// InitializeLcmgrSchema executes init_lcmgr_pgsql.sql to create the lcmgr schema.
// Automatically initializes shared schema first.
func InitializeLcmgrSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_lcmgr_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify lcmgr schema was created
	if !VerifyTableExists(t, db, "lcmgr", "chain_state") {
		return fmt.Errorf("lcmgr.chain_state table was not created")
	}

	return nil
}

// InitializeLcmgrSchemaAndTrezorErc20Tables executes init_trezor_erc20_pgsql.sql to create the trezor erc20 tables.
// Automatically initializes lcmgr schema first (which includes shared schema).
func InitializeLcmgrSchemaAndTrezorErc20Tables(t *testing.T, db *sql.DB) error {
	// Initialize lcmgr schema first (dependency)
	if err := InitializeLcmgrSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize lcmgr schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_trezor_erc20_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify trezor tables were created
	if !VerifyTableExists(t, db, "lcmgr", "trz_erc20_contracts") {
		return fmt.Errorf("lcmgr.trz_erc20_contracts table was not created")
	}

	return nil
}

// InitializeAccmgrSchema executes init_accmgr_pgsql.sql to create the accmgr schema.
// Automatically initializes shared schema first.
func InitializeAccmgrSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_accmgr_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify accmgr schema was created by checking for a key table
	if !VerifyTableExists(t, db, "accmgr", "participants") {
		return fmt.Errorf("accmgr.participants table was not created")
	}

	return nil
}

// InitializeInstrmgrSchema executes init_instrmgr_pgsql.sql to create the instrmgr schema.
// Automatically initializes shared schema first.
func InitializeInstrmgrSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_instrmgr_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify instrmgr schema was created
	if !VerifyTableExists(t, db, "instrmgr", "assets") {
		return fmt.Errorf("instrmgr.assets table was not created")
	}

	return nil
}

// InitializeLaserSchema executes init_laser_pgsql.sql to create the laser schema.
// Automatically initializes shared schema first.
func InitializeLaserSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_laser_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Verify laser schema was created
	if !VerifyTableExists(t, db, "laser", "executors") {
		return fmt.Errorf("laser.executors table was not created")
	}

	return nil
}

// InitializeCsdmsggwSchema executes init_csdmsggw_pgsql.sql to create the csdmsggw schema.
// Automatically initializes shared schema first.
func InitializeCsdmsggwSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_csdmsggw_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Note: Verify with appropriate table for csdmsggw schema
	// Adjust table name based on actual schema structure
	return nil
}

// InitializeMarketmgrSchema executes init_marketmgr_pgsql.sql to create the marketmgr schema.
// Automatically initializes shared schema first.
func InitializeMarketmgrSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_marketmgr_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Note: Verify with appropriate table for marketmgr schema
	return nil
}

// InitializeTraxSchema executes init_trax_pgsql.sql to create the trax schema.
// Automatically initializes shared schema first.
func InitializeTraxSchema(t *testing.T, db *sql.DB) error {
	// Initialize shared schema first (dependency)
	if err := InitializeSharedSchema(t, db); err != nil {
		return fmt.Errorf("failed to initialize shared schema: %w", err)
	}

	projectRoot := FindProjectRoot(t)
	sqlPath := filepath.Join(projectRoot, "deploy/k8s/init/init_trax_pgsql.sql")

	ExecuteSQLFile(t, db, sqlPath)

	// Note: Verify with appropriate table for trax schema
	return nil
}
