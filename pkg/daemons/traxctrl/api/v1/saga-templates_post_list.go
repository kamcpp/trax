package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /saga-templates/list [post]
// @summary List all saga templates with full details
// @schemes
// @tags saga
// @produce json
// @success 200 object listSagaTemplatesResponse
// @failure 500 "internal server error; check server logs"
func postSagaTemplates(c *gin.Context) {
	// Fetch all saga templates from the store
	sagaTemplateList, err := traxStore.ListSagaTemplates(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list saga templates: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	var sagaTemplates []sagaTemplateResponse
	for _, sagaTemplate := range sagaTemplateList {
		templateResp, err := convertSagaTemplateToResponse(c, sagaTemplate)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert saga template %q: %v", sagaTemplate.TemplateId, err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}

		sagaTemplates = append(sagaTemplates, *templateResp)
	}

	resp := listSagaTemplatesResponse{
		SagaTemplates: sagaTemplates,
	}
	c.JSON(200, resp)
}
