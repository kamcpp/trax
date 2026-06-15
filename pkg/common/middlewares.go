package common

import (
	"time"

	"github.com/gin-gonic/gin"
)

func DelayMiddleware(delayMs uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
		c.Next()
	}
}

func RequestAndTraceIdsAttacherMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		present := true
		var requestId string
		requestId = c.GetHeader("x-request-id")
		if len(requestId) == 0 {
			requestId, present = c.GetQuery("request-id")
			if !present {
				requestId = SecureRandomString(32)
			}
			c.Request.Header.Add("x-request-id", requestId)
		}
		if present {
			/* L.Debug(
				fmt.Sprintf("request id is present: %s", requestId),
				F(c)...,
			) */
		} else {
			/* L.Debug(
				fmt.Sprintf("request id is attached: %s", requestId),
				F(c)...,
			) */
		}
		var traceId string
		present = true
		traceId = c.GetHeader("x-trace-id")
		if len(traceId) == 0 {
			traceId, present = c.GetQuery("trace-id")
			if !present {
				traceId = requestId
			}
			c.Request.Header.Add("x-trace-id", traceId)
		}
		/* L.Debug(
			fmt.Sprintf("trace id: %s", traceId),
			F(c)...,
		) */
		c.Next()
	}
}

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		params := gin.LogFormatterParams{}

		params.TimeStamp = time.Now()
		params.Latency = params.TimeStamp.Sub(start)
		params.ClientIP = c.ClientIP()
		params.Method = c.Request.Method
		params.StatusCode = c.Writer.Status()
		params.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		params.BodySize = c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}
		params.Path = path

		/* L.Info("round-trip-http", F(
			c,
			zap.String("sub_sub_component", "gin"),
			zap.String("client_ip", params.ClientIP),
			zap.String("method", params.Method),
			zap.Int("status_code", params.StatusCode),
			zap.Int("body_size", params.BodySize),
			zap.String("path", params.Path),
			zap.Duration("latency", params.Latency),
			zap.Int64("latency_us", params.Latency.Microseconds()),
			zap.String("err_msg", params.ErrorMessage),
		)...) */
	}
}

func ApiKeyCheckerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var apiKey string
		apiKey = c.GetHeader("x-agora-api-key")
		if len(apiKey) == 0 {
			var found bool
			apiKey, found = c.GetQuery("agora-api-key")
			if !found {
				L.Warn("api-key header not found", F(c)...)
				c.JSON(401, gin.H{})
				c.Abort()
				return
			}
		}
		clients := GetClients()
		_, ok := clients[apiKey]
		if !ok {
			L.Warn("no client found", F(c)...)
			c.JSON(401, gin.H{})
			c.Abort()
			return
		}
		c.Next()
	}
}
