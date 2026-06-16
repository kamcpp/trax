package apiv1

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xshyft/trax/pkg/common"
	"github.com/xshyft/trax/pkg/trax"
)

type setDatabaseNameRequest struct {
	DatabaseName string `json:"database_name" binding:"required"`
}

type setDatabaseNameResponse struct {
	Message      string `json:"message"`
	DatabaseName string `json:"database_name"`
	Connected    bool   `json:"connected"`
}

// postSetDatabaseName dynamically switches the database connection
// This endpoint is ONLY available when ENABLE_TESTING_ENDPOINTS=true
//
// @Summary      Set database name (testing only)
// @Description  Dynamically switch to a different database. Only enabled when ENABLE_TESTING_ENDPOINTS=true
// @Tags         experimental/testing
// @Accept       json
// @Produce      json
// @Param        request body setDatabaseNameRequest true "Database name"
// @Success      200 {object} setDatabaseNameResponse
// @Failure      404 {object} map[string]string "Testing endpoints not enabled"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Failed to switch database"
// @Router       /experimental/testing/setdbname [post]
func postSetDatabaseName(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if testing endpoints are enabled
	if os.Getenv("ENABLE_TESTING_ENDPOINTS") != "true" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
		})
		return
	}

	var req setDatabaseNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := fmt.Sprintf("invalid request for setdbname: %v", err)
		common.L.Error(errMsg, common.F(ctx)...)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	common.L.Info(fmt.Sprintf("Switching database to: %s", req.DatabaseName), common.F(ctx)...)

	// Get current POSTGRESQL_CONN_STRING and extract components
	pgsqlURL := os.Getenv("POSTGRESQL_CONN_STRING")
	if pgsqlURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "POSTGRESQL_CONN_STRING not configured",
		})
		return
	}

	// Build new connection string with the new database name
	// Format: postgres://user:password@host:port/dbname?params
	newPgsqlURL := replaceDatabase(pgsqlURL, req.DatabaseName)

	common.L.Info(fmt.Sprintf("Creating new database connection: %s", maskPassword(newPgsqlURL)), common.F(ctx)...)

	// Create new store with new database
	newStore, err := trax.NewPsqlStore(newPgsqlURL)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create new database connection: %v", err)
		common.L.Error(errMsg, common.F(ctx)...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to connect to database: %v", err),
		})
		return
	}

	// Initialize the new store (creates cluster tables if needed)
	if err := newStore.Init(ctx); err != nil {
		errMsg := fmt.Sprintf("failed to initialize new store: %v", err)
		common.L.Error(errMsg, common.F(ctx)...)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to initialize store: %v", err),
		})
		return
	}

	// Save reference to old store
	oldStore := traxStore

	// Replace the global store FIRST (before closing old one)
	// This ensures new requests use the new connection immediately
	traxStore = newStore

	common.L.Info("Recreated traxStore with new database connection", common.F(ctx)...)

	// Wait for connections to stabilize with new database
	common.L.Info("Waiting for connections to stabilize with new database...", common.F(ctx)...)
	time.Sleep(5 * time.Second)
	common.L.Info("Connections stabilized", common.F(ctx)...)

	// Close old store connections after a brief delay
	// This allows any in-flight requests to complete
	if oldStore != nil {
		common.L.Info("Closing old database connections after brief delay...", common.F(ctx)...)
		go func() {
			// Wait for in-flight requests to complete (typical request < 100ms)
			time.Sleep(200 * time.Millisecond)
			if err := oldStore.Close(); err != nil {
				common.L.Warn(fmt.Sprintf("Error closing old database connections: %v", err))
			} else {
				common.L.Info("Old database connections closed successfully")
			}
		}()
	}

	common.L.Info(fmt.Sprintf("Database switched successfully to: %s", req.DatabaseName), common.F(ctx)...)

	c.JSON(http.StatusOK, setDatabaseNameResponse{
		Message:      "database switched successfully",
		DatabaseName: req.DatabaseName,
		Connected:    true,
	})
}

// replaceDatabase replaces the database name in a PostgreSQL connection string
func replaceDatabase(pgsqlURL, newDBName string) string {
	// Handle format: postgres://user:pass@host:port/dbname?params
	parts := strings.Split(pgsqlURL, "/")
	if len(parts) < 4 {
		return pgsqlURL
	}

	// Get the last part which contains dbname and possibly query params
	lastPart := parts[len(parts)-1]

	// Check if there are query parameters
	if strings.Contains(lastPart, "?") {
		paramsParts := strings.Split(lastPart, "?")
		parts[len(parts)-1] = newDBName + "?" + paramsParts[1]
	} else {
		parts[len(parts)-1] = newDBName
	}

	return strings.Join(parts, "/")
}

// maskPassword masks the password in a connection string for logging
func maskPassword(pgsqlURL string) string {
	// Format: postgres://user:password@host:port/dbname
	if !strings.Contains(pgsqlURL, "://") {
		return pgsqlURL
	}

	parts := strings.SplitN(pgsqlURL, "://", 2)
	if len(parts) != 2 {
		return pgsqlURL
	}

	afterScheme := parts[1]
	if !strings.Contains(afterScheme, "@") {
		return pgsqlURL
	}

	userParts := strings.SplitN(afterScheme, "@", 2)
	userInfo := userParts[0]

	if !strings.Contains(userInfo, ":") {
		return pgsqlURL
	}

	userPassParts := strings.SplitN(userInfo, ":", 2)
	return parts[0] + "://" + userPassParts[0] + ":****@" + userParts[1]
}
