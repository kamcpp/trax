package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /clusters [post]
// @summary Create a new cluster
// @schemes
// @tags cluster
// @accept json
// @produce json
// @param request body createClusterRequest true "Cluster creation data"
// @success 201 object clusterResponse
// @failure 400 "bad request; check request body"
// @failure 409 "cluster already exists"
// @failure 500 "internal server error; check server logs"
func postCreateCluster(c *gin.Context) {
	var req createClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := fmt.Sprintf("invalid request body: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	cluster, err := convertCreateRequestToCluster(&req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert request to cluster: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	created, err := traxStore.SaveClusterIdempotently(c, cluster)
	if err != nil {
		errMsg := fmt.Sprintf("failed to save cluster: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	if !created {
		errMsg := fmt.Sprintf("cluster %q already exists", cluster.Id)
		common.L.Warn(errMsg, common.F(c)...)
		c.JSON(409, errMsg)
		return
	}

	resp, err := convertClusterToResponse(cluster)
	if err != nil {
		errMsg := fmt.Sprintf("failed to convert cluster to response: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}
	c.JSON(201, resp)
}
