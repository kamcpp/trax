package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/trax"
)

// @version v1
// @router /saga-step-instances/list [post]
// @summary List saga step instances with full details for a cluster, optionally filtered by saga instance ID
// @schemes
// @tags saga
// @produce json
// @param request body listSagaStepInstancesRequest true "Request body with cluster ID and optional saga instance ID"
// @success 200 object listSagaStepInstancesResponse
// @failure 400 "cluster ID is required"
// @failure 500 "internal server error; check server logs"
func postSagaStepInstances(c *gin.Context) {
	var req listSagaStepInstancesRequest
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

	var sagaStepInstances []*trax.SagaStepInstance
	var err error

	if req.SagaInstanceId != "" {
		// Filter by specific saga instance ID
		sagaStepInstances, err = traxStore.ListSagaStepInstancesBySagaInstanceId(c, req.ClusterId, req.SagaInstanceId)
		if err != nil {
			errMsg := fmt.Sprintf("failed to list saga step instances for saga instance %q in cluster %q: %v", req.SagaInstanceId, req.ClusterId, err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}
	} else {
		// List all saga step instances in the cluster
		sagaStepInstances, err = traxStore.ListSagaStepInstances(c, req.ClusterId)
		if err != nil {
			errMsg := fmt.Sprintf("failed to list saga step instances for cluster %q: %v", req.ClusterId, err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}
	}

	// Order saga steps by their execution sequence (following the linked list structure)
	// This ensures steps are returned in the correct order (1 -> 2 -> 3 -> ...)
	// instead of random database order
	sagaStepInstances = trax.OrderSagaStepsInSequence(sagaStepInstances)

	var instances []sagaStepInstanceResponse
	for _, instance := range sagaStepInstances {
		instanceResp, err := convertSagaStepInstanceToResponse(instance)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert saga step instance to response: %v", err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}
		instances = append(instances, *instanceResp)
	}

	resp := listSagaStepInstancesResponse{
		SagaStepInstances: instances,
	}
	c.JSON(200, resp)
}
