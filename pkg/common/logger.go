package common

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
)

var (
	versionBranch string
	versionHash   string
	L             *zap.Logger
	SubComponent  string
)

var (
	extraFields *map[string]string
)

func InitLogger() {
	versionBranch = os.Getenv("VERSION_BRANCH")
	versionHash = os.Getenv("VERSION_HASH")
	mode := os.Getenv("MODE")
	var config zap.Config
	if mode == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	// Add microsecond precision timestamps (RFC3339Nano)
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	// Set log level from LOG_LEVEL environment variable (default: INFO)
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "INFO" // Default to INFO when not set
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevelStr)); err == nil {
		config.Level = zap.NewAtomicLevelAt(level)
	}

	var err error
	L, err = config.Build(
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}
}

func InitLoggerWithExtraFields(aExtraFields *map[string]string) {
	InitLogger()
	extraFields = aExtraFields
}

func F(ctx context.Context, fields ...zap.Field) []zap.Field {
	fields = append(fields, zap.String("project", "trax"))
	fields = append(fields, zap.String("component", "trax"))
	fields = append(fields, zap.String("sub_component", SubComponent))
	fields = append(fields, zap.String("version_branch", versionBranch))
	fields = append(fields, zap.String("version_hash", versionHash))
	if ctx != nil {
		c, ok := ctx.(*gin.Context)
		if ok {
			fields = append(fields, zap.String("host", c.Request.Host))
			fields = append(fields, zap.String("path", c.Request.URL.Path))
			fields = append(fields, zap.String("method", c.Request.Method))
			requestId := c.GetHeader("x-request-id")
			if len(requestId) > 0 {
				fields = append(fields, zap.String("request_id", requestId))
			} else {
				requestId, ok := c.GetQuery("request-id")
				if ok {
					fields = append(fields, zap.String("request_id", requestId))
				}
			}
			traceId := c.GetHeader("x-trace-id")
			if len(traceId) > 0 {
				fields = append(fields, zap.String("trace_id", traceId))
			} else {
				traceId, ok := c.GetQuery("trace-id")
				if ok {
					fields = append(fields, zap.String("trace_id", traceId))
				}
			}
			apiKey := c.GetHeader("x-agora-api-key")
			if len(apiKey) > 0 {
				fields = append(fields, zap.String("api_key", apiKey))
			} else {
				apiKey, ok := c.GetQuery("api-key")
				if ok {
					fields = append(fields, zap.String("api_key", apiKey))
				}
			}
		}
	}
	if extraFields != nil {
		for key, value := range *extraFields {
			fields = append(fields, zap.String(key, value))
		}
	}
	return fields
}
