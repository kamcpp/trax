package apiv1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @version v1
// @router /saga-instances/{saga_instance_id}/children [post]
// @summary Get child saga instances
// @description Returns direct child saga instances of the given saga instance.
// @schemes
// @tags saga-instances
// @accept json
// @produce json
// @param saga_instance_id path string true "Saga instance ID"
// @param request body getSagaInstanceChildrenRequest true "Request body with cluster_id"
// @success 200 object listSagaInstancesResponse "List of child saga instances"
// @failure 400 "bad request; missing saga instance ID or invalid body"
// @failure 500 "internal server error; check server logs"
func postSagaInstanceChildren(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	if sagaInstanceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sagaInstanceId is required"})
		return
	}

	var req getSagaInstanceChildrenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	children, err := traxStore.GetChildSagaInstances(c, req.ClusterId, sagaInstanceId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []sagaInstanceResponse
	for _, child := range children {
		resp, err := convertSagaInstanceToResponse(child)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		responses = append(responses, *resp)
	}

	if responses == nil {
		responses = []sagaInstanceResponse{}
	}

	c.JSON(http.StatusOK, listSagaInstancesResponse{
		SagaInstances: responses,
	})
}
