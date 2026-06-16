package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /saga-step-instances/{sagaStepInstanceId} [post]
// @summary Get saga step instance by ID
// @schemes
// @tags saga
// @produce json
// @param sagaStepInstanceId path string true "Saga Step Instance ID"
// @param request body getSagaStepInstanceRequest true "Request body with cluster ID"
// @success 200 object sagaStepInstanceResponse
// @failure 400 "saga step instance ID or cluster ID is required"
// @failure 404 "saga step instance not found"
// @failure 500 "internal server error; check server logs"
func postSagaStepInstance(c *gin.Context) {
	sagaStepInstanceId := c.Param("sagaStepInstanceId")
	if sagaStepInstanceId == "" {
		errMsg := "saga step instance ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	var req getSagaStepInstanceRequest
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

	// Get step instance by instance ID
	sagaStepInstance, err := traxStore.GetSagaStepInstance(c, req.ClusterId, sagaStepInstanceId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get saga step instance %q: %v", sagaStepInstanceId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(404, errMsg)
		return
	}

	instanceResp, err := convertSagaStepInstanceToResponse(sagaStepInstance)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert saga step instance to response: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}
	c.JSON(200, instanceResp)
}
