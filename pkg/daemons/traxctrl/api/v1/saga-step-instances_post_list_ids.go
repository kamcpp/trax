package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /saga-step-instances/list/ids [post]
// @summary List all saga step instance IDs for a cluster
// @schemes
// @tags saga
// @produce json
// @param request body listSagaStepInstanceIdsRequest true "Request body with cluster ID"
// @success 200 object listSagaStepInstanceIdsResponse
// @failure 400 "cluster ID is required"
// @failure 500 "internal server error; check server logs"
func postSagaStepInstanceIds(c *gin.Context) {
	var req listSagaStepInstanceIdsRequest
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

	instanceIds, err := traxStore.ListSagaStepInstanceIds(c, req.ClusterId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list saga step instance IDs for cluster %q: %v", req.ClusterId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	resp := listSagaStepInstanceIdsResponse{
		SagaStepInstanceIds: instanceIds,
	}
	c.JSON(200, resp)
}
