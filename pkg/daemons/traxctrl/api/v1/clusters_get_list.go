package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/xshyft/trax/pkg/common"
)

// @version v1
// @router /clusters/list [get]
// @summary List all clusters with full details
// @schemes
// @tags cluster
// @produce json
// @success 200 object listClustersResponse
// @failure 500 "internal server error; check server logs"
func getClusterList(c *gin.Context) {
	clusters, err := traxStore.ListClusters(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list clusters: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	var clusterResponses []clusterResponse
	for _, cluster := range clusters {
		clusterResp, err := convertClusterToResponse(cluster)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert cluster to response: %v", err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}
		clusterResponses = append(clusterResponses, *clusterResp)
	}

	resp := listClustersResponse{
		Clusters: clusterResponses,
	}
	c.JSON(200, resp)
}
