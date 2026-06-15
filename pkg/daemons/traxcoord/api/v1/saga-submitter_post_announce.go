package apiv1

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/trax"
)

// @version v1
// @router /saga-submitter/announce [post]
// @summary Registers a new saga submitter
// @schemes
// @tags saga
// @produce json
// @param request body trax.PostAnnounceSagaSubmitterRequest true "the request"
// @success 200 object trax.PostAnnounceSagaSubmitterResponse
// @failure 500 "internal server error; check server logs"
func postAnnounceSagaSubmitter(c *gin.Context) {
	var request trax.PostAnnounceSagaSubmitterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		// TODO(kam): expose error details only in debug mode
		errMsg := fmt.Sprintf("failed to bind request: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}
	clusterIds, nodeNamesPerClusterMap, err :=
		traxCoordinatorService.AnnounceSagaSubmitter(c, request.SagaSubmitterId)
	if err != nil {
		// TODO(kam): expose error details only in debug mode
		errMsg := fmt.Sprintf("failed to announce saga submitter %q: %v", request.SagaSubmitterId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}
	resp := trax.PostAnnounceSagaSubmitterResponse{
		ClusterIds:             clusterIds,
		NodeNamesPerClusterMap: nodeNamesPerClusterMap,
	}
	c.JSON(200, resp)
}
