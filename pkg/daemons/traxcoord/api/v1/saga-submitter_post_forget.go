package apiv1

import (
	"github.com/gin-gonic/gin"
)

// @version v1
// @router /saga-submitter/forget [post]
// @summary Forget/reset the saga submitter
// @schemes
// @tags saga-submitter
// @produce json
// @success 200 "ok but not supported"
func postForgetSagaSubmitter(c *gin.Context) {
	// NOTE: For now, we don't support forgetting a saga submitter
	c.JSON(200, "ok but not supported")
}
