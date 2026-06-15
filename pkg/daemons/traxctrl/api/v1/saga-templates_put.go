package apiv1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/trax"
)

// @version v1
// @router /saga-templates/{sagaTemplateId} [put]
// @summary Update an existing saga template
// @tags saga
// @accept json
// @produce json
// @param sagaTemplateId path string true "Saga Template ID"
// @param body body updateSagaTemplateRequest true "Template update payload"
// @success 200 "saga template updated"
// @failure 400 "invalid request"
// @failure 404 "saga template not found"
// @failure 500 "internal server error"
func putSagaTemplate(c *gin.Context) {
	sagaTemplateId := c.Param("sagaTemplateId")
	if sagaTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga template ID is required"})
		return
	}

	var req updateSagaTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sagaTemplate := &trax.SagaTemplate{
		TemplateId:          sagaTemplateId,
		DisplayName:         req.DisplayName,
		Description:         req.Description,
		Labels:              req.Labels,
		Tags:                req.Tags,
		Metadata:            req.Metadata,
		SagaStepTemplateIds: req.SagaStepTemplateIds,
	}

	if err := traxStore.UpdateSagaTemplate(c, sagaTemplate); err != nil {
		errMsg := fmt.Sprintf("failed to update saga template %q: %v", sagaTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("saga template '%s' updated", sagaTemplateId)})
}
