package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
)

// @version v1
// @router /clusters/{clusterId} [delete]
// @summary Delete cluster by ID
// @schemes
// @tags cluster
// @produce json
// @param clusterId path string true "Cluster ID"
// @success 204 "cluster deleted successfully"
// @failure 404 "cluster not found"
// @failure 500 "internal server error; check server logs"
func deleteCluster(c *gin.Context) {
	clusterId := c.Param("clusterId")
	if clusterId == "" {
		errMsg := "cluster ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	err := traxStore.DeleteCluster(c, clusterId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to delete cluster %q: %v", clusterId, err)
		common.L.Error(errMsg, common.F(c)...)
		if err.Error() == "cluster not found" {
			c.JSON(404, errMsg)
		} else {
			c.JSON(500, errMsg)
		}
		return
	}

	c.Status(204)
}
