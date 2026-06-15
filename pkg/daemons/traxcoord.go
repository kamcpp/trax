package daemons

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/cache"
	"github.com/kamcpp/trax/pkg/common"
	apiv1 "github.com/kamcpp/trax/pkg/daemons/traxcoord/api/v1"
	"github.com/kamcpp/trax/pkg/mq"
	"github.com/kamcpp/trax/pkg/trax"
)

func RunTraxCoordinator() {

	// TODO(kam): define the context
	ctx := context.Background()

	common.SubComponent = "traxcoord"
	common.InitLogger()

	if os.Getenv("SU_MODE") == "active" {
		common.L.Warn("!!! SU mode is active !!!", common.F(ctx)...)
	}

	affinityGroup := os.Getenv("TRAX_COORDINATOR_AFFINITY_GROUP")
	if len(affinityGroup) == 0 {
		panic("TRAX_COORDINATOR_AFFINITY_GROUP is not set")
	}

	pgsqlUrl := os.Getenv("POSTGRESQL_CONN_STRING")
	if len(pgsqlUrl) == 0 {
		panic("POSTGRESQL_CONN_STRING is not set")
	}

	cache.Init(ctx)
	mq.Init(ctx)

	mqClient := trax.NewRabbitMQClient()

	common.L.Debug("Using PostgreSQL store ...", common.F(ctx)...)
	traxStore, err := trax.NewPsqlStore(pgsqlUrl)
	if err != nil {
		panic(fmt.Sprintf("failed to create trax store: %v", err))
	}
	common.L.Debug("Initializing PostgreSQL store ...", common.F(ctx)...)
	err = traxStore.Init(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize trax store: %v", err))
	}
	common.L.Debug("PostgreSQL store initialized", common.F(ctx)...)

	// Enable LISTEN/NOTIFY for event-driven saga processing
	// This dramatically reduces step-to-step latency from 500ms-3s polling to near-instant
	listenChannel := "trax_saga_events"
	if err := traxStore.Listen(ctx, listenChannel); err != nil {
		common.L.Warn(fmt.Sprintf("Failed to enable LISTEN/NOTIFY on channel '%s': %v. Falling back to polling.", listenChannel, err), common.F(ctx)...)
	} else {
		common.L.Info(fmt.Sprintf("LISTEN/NOTIFY enabled on channel '%s' for event-driven saga processing", listenChannel), common.F(ctx)...)
	}

	// Enable LISTEN/NOTIFY for template hot-reload
	templateChannel := "trax_template_events"
	if err := traxStore.Listen(ctx, templateChannel); err != nil {
		common.L.Warn(fmt.Sprintf("Failed to enable LISTEN/NOTIFY on channel '%s': %v. Template changes will rely on polling.", templateChannel, err), common.F(ctx)...)
	} else {
		common.L.Info(fmt.Sprintf("LISTEN/NOTIFY enabled on channel '%s' for template hot-reload", templateChannel), common.F(ctx)...)
	}

	traxCoordinator := trax.NewSagaCoordinator(
		mqClient, traxStore, affinityGroup)

	go traxCoordinator.Start(ctx)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(common.GinLoggerMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Origin", "Accept" /* "X-Agora-Api-Key" */},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(common.RequestAndTraceIdsAttacherMiddleware())

	apiv1.Init(r, traxCoordinator)

	go r.Run("0.0.0.0:17201") // listen and serve on 0.0.0.0:17201

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	common.L.Warn("Received SIGTERM signal", common.F(ctx)...)
	common.L.Info("Bye", common.F(ctx)...)
}
