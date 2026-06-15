package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type contextKey string

const (
	TraceIdHeaderName contextKey = "x-trace-id"
	TraceIdLength     int        = 32
)

// TraceIdInterceptor is a gRPC unary interceptor that ensures every request has a trace ID
func TraceIdInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var traceId string
	headerMissing := false

	// Extract metadata from incoming context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		headerMissing = true
	} else {
		// Check for trace ID in metadata
		traceIds := md.Get(string(TraceIdHeaderName))
		if len(traceIds) == 0 || traceIds[0] == "" {
			headerMissing = true
		} else {
			traceId = traceIds[0]
		}
	}

	// Generate new trace ID if missing and log warning
	if headerMissing {
		traceId = SecureRandomString(TraceIdLength)
		L.Warn("x-trace-id header missing, assigned new trace id: "+traceId, F(ctx)...)
	}

	// Add trace ID to context for downstream processing
	ctx = context.WithValue(ctx, TraceIdHeaderName, traceId)

	// Call the handler
	resp, err := handler(ctx, req)

	// Set the trace ID in the outgoing response headers
	if err == nil {
		grpc.SetHeader(ctx, metadata.Pairs(string(TraceIdHeaderName), traceId))
	}

	return resp, err
}

// compactHeaders converts metadata to a compact string representation
func compactHeaders(md metadata.MD) string {
	if len(md) == 0 {
		return "{}"
	}

	var parts []string
	for key, values := range md {
		if len(values) == 1 {
			parts = append(parts, fmt.Sprintf("%s:%s", key, values[0]))
		} else {
			parts = append(parts, fmt.Sprintf("%s:[%s]", key, strings.Join(values, ",")))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, " "))
}

// compactArgs converts request arguments to a compact string representation
func compactArgs(req interface{}) string {
	if req == nil {
		return "{}"
	}

	// Try to marshal to JSON for compact representation
	if jsonBytes, err := json.Marshal(req); err == nil {
		jsonStr := string(jsonBytes)
		// Limit length to keep logs readable
		if len(jsonStr) > 200 {
			return jsonStr[:197] + "..."
		}
		return jsonStr
	}

	// Fallback to string representation
	reqStr := fmt.Sprintf("%+v", req)
	if len(reqStr) > 200 {
		return reqStr[:197] + "..."
	}
	return reqStr
}

// RequestLoggingInterceptor is a gRPC unary interceptor that logs every incoming request
func RequestLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Get client peer information
	var clientAddr string
	if p, ok := peer.FromContext(ctx); ok {
		clientAddr = p.Addr.String()
	}

	// Get trace ID from context (set by TraceIdInterceptor)
	var traceId string
	if ctxTraceId, ok := ctx.Value(TraceIdHeaderName).(string); ok {
		traceId = ctxTraceId
	}

	// Get headers from incoming metadata
	var headersStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		headersStr = compactHeaders(md)
	} else {
		headersStr = "{}"
	}

	// Get compact request arguments
	argsStr := compactArgs(req)

	// Log incoming request
	L.Info("grpc start",
		append(F(ctx),
			zap.String("method", info.FullMethod),
			zap.String("client_addr", clientAddr),
			zap.String("trace_id", traceId),
			zap.String("headers", headersStr),
			zap.String("args", argsStr),
		)...)

	// Call the handler
	resp, err := handler(ctx, req)

	// Calculate duration
	duration := time.Since(start)

	// Log request completion
	if err != nil {
		L.Error("grpc error",
			append(F(ctx),
				zap.String("method", info.FullMethod),
				zap.String("client_addr", clientAddr),
				zap.String("trace_id", traceId),
				zap.Int64("dur_ms", duration.Milliseconds()),
				zap.String("headers", headersStr),
				zap.String("args", argsStr),
				zap.String("err", err.Error()),
			)...)
	} else {
		/* L.Info("grpc ok",
		append(F(ctx),
			zap.String("method", info.FullMethod),
			zap.String("client_addr", clientAddr),
			zap.String("trace_id", traceId),
			zap.Int64("dur_ms", duration.Milliseconds()),
			zap.String("headers", headersStr),
			zap.String("args", argsStr),
		)...) */
	}

	return resp, err
}

// ChainUnaryInterceptors chains multiple unary interceptors into one
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	if len(interceptors) == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}
	if len(interceptors) == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		currHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			innerHandler := currHandler
			interceptor := interceptors[i]
			currHandler = func(ctx context.Context, req interface{}) (interface{}, error) {
				return interceptor(ctx, req, info, innerHandler)
			}
		}
		return currHandler(ctx, req)
	}
}
