package apiv1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /saga-step-templates/{sagaStepTemplateId} [delete]
// @summary Delete a saga step template
// @tags saga
// @produce json
// @param sagaStepTemplateId path string true "Saga Step Template ID"
// @success 200 "saga step template deleted"
// @failure 400 "invalid request"
// @failure 404 "saga step template not found"
// @failure 500 "internal server error"
func deleteSagaStepTemplate(c *gin.Context) {
	sagaStepTemplateId := c.Param("sagaStepTemplateId")
	if sagaStepTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga step template ID is required"})
		return
	}

	if err := traxStore.DeleteSagaStepTemplate(c, sagaStepTemplateId); err != nil {
		errMsg := fmt.Sprintf("failed to delete saga step template %q: %v", sagaStepTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("saga step template '%s' deleted", sagaStepTemplateId)})
}
