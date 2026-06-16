package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /saga-templates/{sagaTemplateId} [post]
// @summary Get saga template by ID with steps
// @schemes
// @tags saga
// @produce json
// @param sagaTemplateId path string true "Saga Template ID"
// @success 200 object sagaTemplateResponse
// @failure 404 "saga template not found"
// @failure 500 "internal server error; check server logs"
func postSagaTemplate(c *gin.Context) {
	sagaTemplateId := c.Param("sagaTemplateId")
	if sagaTemplateId == "" {
		errMsg := "saga template ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	// Get specific template
	sagaTemplate, err := traxStore.GetSagaTemplate(c, sagaTemplateId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get saga template %q: %v", sagaTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(404, errMsg)
		return
	}

	templateResp, err := convertSagaTemplateToResponse(c, sagaTemplate)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert saga template %q: %v", sagaTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	c.JSON(200, templateResp)
}
