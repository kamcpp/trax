package apiv1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
	"github.com/xshyft/trax/pkg/trax"
)

// @version v1
// @router /saga-step-templates/{sagaStepTemplateId} [put]
// @summary Update an existing saga step template
// @tags saga
// @accept json
// @produce json
// @param sagaStepTemplateId path string true "Saga Step Template ID"
// @param body body updateSagaStepTemplateRequest true "Step template update payload"
// @success 200 "saga step template updated"
// @failure 400 "invalid request"
// @failure 404 "saga step template not found"
// @failure 500 "internal server error"
func putSagaStepTemplate(c *gin.Context) {
	sagaStepTemplateId := c.Param("sagaStepTemplateId")
	if sagaStepTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga step template ID is required"})
		return
	}

	var req updateSagaStepTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sagaStepTemplate := &trax.SagaStepTemplate{
		TemplateId:     sagaStepTemplateId,
		SagaTemplateId: req.SagaTemplateId,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		Labels:         req.Labels,
		Tags:           req.Tags,
		Metadata:       req.Metadata,
	}

	if err := traxStore.UpdateSagaStepTemplate(c, sagaStepTemplate); err != nil {
		errMsg := fmt.Sprintf("failed to update saga step template %q: %v", sagaStepTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("saga step template '%s' updated", sagaStepTemplateId)})
}
