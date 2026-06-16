package apiv1

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /experimental/testing/create-smoke-template [post]
// @summary Create smoke test saga template (testing only)
// @schemes
// @tags testing
// @produce json
// @success 200 object createSmokeTemplateResponse
// @failure 404 "endpoint not found (testing endpoints not enabled)"
// @failure 500 "internal server error; check server logs"
// @failure 501 "not implemented"
func postCreateSmokeTemplate(c *gin.Context) {
	ctx := c.Request.Context()

	if os.Getenv("ENABLE_TESTING_ENDPOINTS") != "true" {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	common.L.Info("Smoke test template endpoint called but not implemented", common.F(ctx)...)

	// NOTE: Smoke test templates should be created via trax CLI or direct database operations
	// This endpoint is deprecated
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented - use trax CLI instead"})
}

type createSmokeTemplateResponse struct {
	Message string `json:"message"`
}
