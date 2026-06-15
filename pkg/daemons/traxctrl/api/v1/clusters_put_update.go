package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /clusters/{clusterId} [put]
// @summary Update cluster by ID
// @schemes
// @tags cluster
// @accept json
// @produce json
// @param clusterId path string true "Cluster ID"
// @param request body updateClusterRequest true "Cluster update data"
// @success 200 object clusterResponse
// @failure 400 "bad request; check request body"
// @failure 404 "cluster not found"
// @failure 500 "internal server error; check server logs"
func putUpdateCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")
	if clusterId == "" {
		errMsg := "cluster ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	var req updateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := fmt.Sprintf("invalid request body: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	cluster, err := convertUpdateRequestToCluster(clusterId, &req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert request to cluster: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	err = traxStore.UpdateCluster(c, cluster)
	if err != nil {
		errMsg := fmt.Sprintf("failed to update cluster %q: %v", clusterId, err)
		common.L.Error(errMsg, common.F(c)...)
		if err.Error() == "cluster not found" {
			c.JSON(404, errMsg)
		} else {
			c.JSON(500, errMsg)
		}
		return
	}

	resp, err := convertClusterToResponse(cluster)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert cluster to response: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}
	c.JSON(200, resp)
}
