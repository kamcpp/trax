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
	apiv1 "github.com/kamcpp/trax/pkg/daemons/traxctrl/api/v1"
	"github.com/kamcpp/trax/pkg/mq"
	"github.com/kamcpp/trax/pkg/trax"
)

func RunTraxCtrl(useInMemory bool) {
	// TODO(kam): define the context
	ctx := context.Background()

	common.SubComponent = "traxctrl"
	common.InitLogger()

	if os.Getenv("SU_MODE") == "active" {
		common.L.Warn("!!! SU mode is active !!!", common.F(ctx)...)
	}

	cache.Init(ctx)
	mq.Init(ctx)

	var traxStore trax.Store
	var err error

	if useInMemory {
		common.L.Info("Using in-memory store", common.F(ctx)...)
		traxStore = trax.NewInMemoryStore()
	} else {
		pgsqlUrl := os.Getenv("POSTGRESQL_CONN_STRING")
		if len(pgsqlUrl) == 0 {
			panic("POSTGRESQL_CONN_STRING is not set (required when not using --in-memory-store flag)")
		}

		common.L.Info("Using PostgreSQL store", common.F(ctx)...)
		traxStore, err = trax.NewPsqlStore(pgsqlUrl)
		if err != nil {
			panic(fmt.Sprintf("failed to create trax store: %v", err))
		}
	}

	err = traxStore.Init(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize trax store: %v", err))
	}

	// NOTE: Saga templates are now loaded via SQL files in deploy/k8s/init/[namespace]/min/trax.sql
	// Use: ./deploy data min-records --cluster-id <cluster> --ns <namespace>

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(common.GinLoggerMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Origin", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(common.RequestAndTraceIdsAttacherMiddleware())

	apiv1.Init(r, traxStore)

	go r.Run("0.0.0.0:17202") // listen and serve on 0.0.0.0:17202

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	common.L.Warn("Received SIGTERM signal", common.F(ctx)...)
	common.L.Info("Bye", common.F(ctx)...)
}
