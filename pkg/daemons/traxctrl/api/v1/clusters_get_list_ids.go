package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /clusters/list/ids [get]
// @summary List all cluster IDs
// @schemes
// @tags cluster
// @produce json
// @success 200 object listClusterIdsResponse
// @failure 500 "internal server error; check server logs"
func getClusters(c *gin.Context) {
	clusterIds, err := traxStore.ListClusterIds(c)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list cluster IDs: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	if clusterIds == nil {
		clusterIds = []string{}
	}

	resp := listClusterIdsResponse{
		ClusterIds: clusterIds,
	}
	c.JSON(200, resp)
}
