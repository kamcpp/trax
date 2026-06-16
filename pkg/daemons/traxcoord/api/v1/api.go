package apiv1

import (
	"os"

	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/xshyft/trax/gen-docs/traxcoord/v1"
	"github.com/xshyft/trax/pkg/trax"
)

const (
	ApiV1UriPrefix = "/api/v1"
)

var (
	traxCoordinatorService trax.SagaCoordinator
)

func Init(r *gin.Engine, traxCoordinator trax.SagaCoordinator) {

	traxCoordinatorService = traxCoordinator

	r.GET(
		ApiV1UriPrefix+"/health",
		func(c *gin.Context) {
			if traxCoordinatorService.IsReady(c.Request.Context()) {
				c.JSON(200, "ok")
			} else {
				c.JSON(503, gin.H{"ready": false})
			}
		},
	)

	docs.SwaggerInfoTraxCoordinatorV1.Host = os.Getenv("V1_SWAGGER_HOST")
	docs.SwaggerInfoTraxCoordinatorV1.BasePath = ApiV1UriPrefix
	r.GET(
		ApiV1UriPrefix+"/swagger/*any",
		ginSwagger.WrapHandler(swaggerfiles.Handler, func(c *ginSwagger.Config) {
			c.InstanceName = "TraxCoordinatorV1"
		}),
	)
	r.POST(ApiV1UriPrefix+"/saga-submitter/announce", postAnnounceSagaSubmitter)
	r.POST(ApiV1UriPrefix+"/saga-submitter/forget", postForgetSagaSubmitter)

	// Experimental/testing endpoints (only enabled when ENABLE_TESTING_ENDPOINTS=true)
	r.POST(ApiV1UriPrefix+"/experimental/testing/setdbname", postSetDatabaseName)
}
