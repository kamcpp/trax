package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /saga-instances/{sagaInstanceId} [post]
// @summary Get saga instance by ID
// @schemes
// @tags saga
// @produce json
// @param sagaInstanceId path string true "Saga Instance ID"
// @param request body getSagaInstanceRequest true "Request body with cluster ID"
// @success 200 object sagaInstanceResponse
// @failure 400 "saga instance ID or cluster ID is required"
// @failure 404 "saga instance not found"
// @failure 500 "internal server error; check server logs"
func postSagaInstance(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	if sagaInstanceId == "" {
		errMsg := "saga instance ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	var req getSagaInstanceRequest
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

	sagaInstance, err := traxStore.GetSagaInstance(c, req.ClusterId, sagaInstanceId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get saga instance %q: %v", sagaInstanceId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(404, errMsg)
		return
	}

	instanceResp, err := convertSagaInstanceToResponse(sagaInstance)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert saga instance to response: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}
	c.JSON(200, instanceResp)
}
