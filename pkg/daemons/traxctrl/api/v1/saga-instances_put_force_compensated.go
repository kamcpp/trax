package apiv1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kamcpp/trax/pkg/common"
)

// forceMarkSagaCompensatedRequest is the JSON body for the
// "operator override unblock" endpoint. The reason is required —
// it lands on the saga's compensation_reason column for audit so
// every manual unblock has a paper trail.
type forceMarkSagaCompensatedRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
	Reason    string `json:"reason"     binding:"required"`
}

// @version v1
// @router /saga-instances/{sagaInstanceId}/force-compensated [put]
// @summary Force-mark a BLOCKED saga as COMPENSATED (operator override)
// @description Operator-only escape hatch for sagas wedged in BLOCKED
// @description because a compensation step can't recover on its own
// @description (e.g. trying to delete a row another path already
// @description deleted). Refuses to act on sagas that aren't BLOCKED so
// @description the override can't accidentally short-circuit a healthy
// @description saga. The reason lands on the saga's compensation_reason
// @description column for audit.
// @schemes
// @tags saga
// @accept json
// @produce json
// @param sagaInstanceId path string true "Saga Instance ID"
// @param request body forceMarkSagaCompensatedRequest true "Cluster id + audit reason"
// @success 200 {object} map[string]string "saga marked compensated"
// @failure 400 "invalid request body or saga not in BLOCKED state"
// @failure 404 "saga not found"
// @failure 500 "internal server error"
func putForceMarkSagaCompensated(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	if sagaInstanceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sagaInstanceId is required"})
		return
	}

	var req forceMarkSagaCompensatedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("invalid request body: %v", err),
		})
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	if err := traxStore.ForceMarkSagaCompensated(c, req.ClusterId, sagaInstanceId, req.Reason); err != nil {
		errMsg := fmt.Sprintf("failed to force-mark saga compensated: %v", err)
		common.L.Error(errMsg, common.F(c,
			zap.String("saga_instance_id", sagaInstanceId),
			zap.String("cluster_id", req.ClusterId),
			zap.String("reason", req.Reason),
		)...)
		// State-mismatch / not-BLOCKED comes back from the store — surface
		// as 400 so the UI shows the explanation; "not found" surfaces
		// from the store as well; anything else is an upstream bug.
		if strings.Contains(err.Error(), "not in BLOCKED state") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to force-mark saga compensated"})
		return
	}

	common.L.Info("force-marked saga as compensated",
		common.F(c,
			zap.String("saga_instance_id", sagaInstanceId),
			zap.String("cluster_id", req.ClusterId),
			zap.String("reason", req.Reason),
		)...)
	c.JSON(http.StatusOK, gin.H{
		"saga_instance_id": sagaInstanceId,
		"cluster_id":       req.ClusterId,
		"state":            "COMPENSATED",
		"message":          "saga force-marked as compensated",
	})
}
