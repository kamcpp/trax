package apiv1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /saga-templates/{sagaTemplateId} [delete]
// @summary Delete a saga template and its associated step templates
// @tags saga
// @produce json
// @param sagaTemplateId path string true "Saga Template ID"
// @success 200 "saga template deleted"
// @failure 400 "invalid request"
// @failure 404 "saga template not found"
// @failure 500 "internal server error"
func deleteSagaTemplate(c *gin.Context) {
	sagaTemplateId := c.Param("sagaTemplateId")
	if sagaTemplateId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga template ID is required"})
		return
	}

	if err := traxStore.DeleteSagaTemplate(c, sagaTemplateId); err != nil {
		errMsg := fmt.Sprintf("failed to delete saga template %q: %v", sagaTemplateId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("saga template '%s' deleted", sagaTemplateId)})
}
