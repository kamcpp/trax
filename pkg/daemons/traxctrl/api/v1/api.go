package apiv1

import (
	"os"

	"github.com/gin-gonic/gin"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/xshyft/trax/gen-docs/traxctrl/v1"
	"github.com/xshyft/trax/pkg/trax"
)

const (
	ApiV1UriPrefix = "/api/v1"
)

var (
	traxStore trax.Store
)

func Init(r *gin.Engine, store trax.Store) {

	traxStore = store

	r.GET(
		ApiV1UriPrefix+"/health",
		func(c *gin.Context) {
			// common.L.Debug("health ok", common.F(c)...)
			c.JSON(200, "ok")
		},
	)

	docs.SwaggerInfoTraxCtrlV1.Host = os.Getenv("V1_SWAGGER_HOST")
	docs.SwaggerInfoTraxCtrlV1.BasePath = ApiV1UriPrefix
	r.GET(
		ApiV1UriPrefix+"/swagger/*any",
		ginSwagger.WrapHandler(swaggerfiles.Handler, func(c *ginSwagger.Config) {
			c.InstanceName = "TraxCtrlV1"
		}),
	)
	// Register more specific routes first to avoid path parameter conflicts
	r.POST(ApiV1UriPrefix+"/saga-templates/list/ids", postSagaTemplateIds)
	r.POST(ApiV1UriPrefix+"/saga-templates/list", postSagaTemplates)
	r.POST(ApiV1UriPrefix+"/saga-templates/:sagaTemplateId", postSagaTemplate)
	r.PUT(ApiV1UriPrefix+"/saga-templates/:sagaTemplateId", putSagaTemplate)
	r.DELETE(ApiV1UriPrefix+"/saga-templates/:sagaTemplateId", deleteSagaTemplate)

	// Saga step template management endpoints
	r.PUT(ApiV1UriPrefix+"/saga-step-templates/:sagaStepTemplateId", putSagaStepTemplate)
	r.DELETE(ApiV1UriPrefix+"/saga-step-templates/:sagaStepTemplateId", deleteSagaStepTemplate)

	// Saga instance endpoints
	r.POST(ApiV1UriPrefix+"/saga-instances/list/ids", postSagaInstanceIds)
	r.POST(ApiV1UriPrefix+"/saga-instances/list", postSagaInstances)
	r.POST(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/children", postSagaInstanceChildren)
	r.POST(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/tree", postSagaInstanceTree)
	r.POST(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/annexes", postSagaAnnex)
	r.GET(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/annexes", getSagaAnnexes)
	r.GET(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/annexes/:annexIid", getSagaAnnexBytes)
	r.POST(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId", postSagaInstance)
	r.PUT(ApiV1UriPrefix+"/saga-instances/:sagaInstanceId/force-compensated", putForceMarkSagaCompensated)

	// Saga step instance endpoints
	r.POST(ApiV1UriPrefix+"/saga-step-instances/list/ids", postSagaStepInstanceIds)
	r.POST(ApiV1UriPrefix+"/saga-step-instances/list", postSagaStepInstances)
	r.POST(ApiV1UriPrefix+"/saga-step-instances/:sagaStepInstanceId", postSagaStepInstance)

	// Cluster endpoints
	r.POST(ApiV1UriPrefix+"/clusters", postCreateCluster)
	r.GET(ApiV1UriPrefix+"/clusters/list/ids", getClusters)
	r.GET(ApiV1UriPrefix+"/clusters/list", getClusterList)
	r.GET(ApiV1UriPrefix+"/clusters/:clusterId", getCluster)
	r.PUT(ApiV1UriPrefix+"/clusters/:clusterId", putUpdateCluster)
	r.DELETE(ApiV1UriPrefix+"/clusters/:clusterId", deleteCluster)

	// Experimental/testing endpoints (only enabled when ENABLE_TESTING_ENDPOINTS=true)
	r.POST(ApiV1UriPrefix+"/experimental/testing/setdbname", postSetDatabaseName)
	r.POST(ApiV1UriPrefix+"/experimental/testing/create-smoke-template", postCreateSmokeTemplate)
}
