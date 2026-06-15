package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @version v1
// @router /saga-instances/{saga_instance_id}/tree [post]
// @summary Get saga instance tree structure
// @description Returns the full saga hierarchy tree rooted at the given saga instance.
// @schemes
// @tags saga-instances
// @accept json
// @produce json
// @param saga_instance_id path string true "Saga instance ID"
// @param request body getSagaInstanceTreeRequest true "Request body with cluster_id"
// @success 200 object sagaInstanceTreeResponse "Saga instance tree"
// @failure 400 "bad request; missing saga instance ID or invalid body"
// @failure 500 "internal server error; check server logs"
func postSagaInstanceTree(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	if sagaInstanceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sagaInstanceId is required"})
		return
	}

	var req getSagaInstanceTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hierarchy, err := traxStore.GetSagaHierarchy(c, req.ClusterId, sagaInstanceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []sagaInstanceResponse
	for _, saga := range hierarchy {
		resp, err := convertSagaInstanceToResponse(saga)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		responses = append(responses, *resp)
	}

	if responses == nil {
		responses = []sagaInstanceResponse{}
	}

	c.JSON(http.StatusOK, sagaInstanceTreeResponse{
		SagaInstances: responses,
	})
}
