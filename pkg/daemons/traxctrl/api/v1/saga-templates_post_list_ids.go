package apiv1

import (
	"github.com/gin-gonic/gin"
)

// @version v1
// @router /saga-templates/list/ids [post]
// @summary List all saga template IDs
// @schemes
// @tags saga
// @produce json
// @success 200 object listSagaTemplateIdsResponse
// @failure 500 "internal server error; check server logs"
func postSagaTemplateIds(c *gin.Context) {
	// TODO(kam): Implement actual listing from store
	// For now, return the templates we know exist
	templateIds := []string{"new_account_under_participant"}

	resp := listSagaTemplateIdsResponse{
		SagaTemplateIds: templateIds,
	}
	c.JSON(200, resp)
}
