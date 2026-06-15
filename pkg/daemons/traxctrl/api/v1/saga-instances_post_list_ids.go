package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /saga-instances/list/ids [post]
// @summary List all saga instance IDs for a cluster
// @schemes
// @tags saga
// @produce json
// @param request body listSagaInstanceIdsRequest true "Request body with cluster ID"
// @success 200 object listSagaInstanceIdsResponse
// @failure 400 "cluster ID is required"
// @failure 500 "internal server error; check server logs"
func postSagaInstanceIds(c *gin.Context) {
	var req listSagaInstanceIdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := fmt.Sprintf("invalid request body: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	if req.ClusterId == "" {
		errMsg := "cluster ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	instanceIds, err := traxStore.ListSagaInstanceIds(c, req.ClusterId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list saga instance IDs for cluster %q: %v", req.ClusterId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	resp := listSagaInstanceIdsResponse{
		SagaInstanceIds: instanceIds,
	}
	c.JSON(200, resp)
}
