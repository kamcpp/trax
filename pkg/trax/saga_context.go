package trax

import (
	"context"
	"fmt"
	"time"

	"github.com/xshyft/trax/pkg/common"
)

// SagaContext provides saga execution context to IdempotentService implementations.
// It is injected by the executor framework via context.WithValue and carries the parent
// saga identity, enabling automatic parent-child registration when spawning sub-sagas.
type SagaContext interface {
	// ParentSagaInstanceId returns the current saga instance ID
	// (which becomes the parent for any sub-saga spawned through this context)
	ParentSagaInstanceId() string

	// ParentSagaStepInstanceId returns the current saga step instance ID
	ParentSagaStepInstanceId() string

	// RootSagaInstanceId returns the root saga instance ID in the hierarchy
	RootSagaInstanceId() string

	// SagaDepth returns the current depth in the saga hierarchy (0 = top-level)
	SagaDepth() int

	// ClusterId returns the cluster this saga is running in
	ClusterId() string

	// SpawnSubSaga submits a sub-saga with automatic parent-child registration,
	// waits for completion, and returns the result. Handles:
	//  - Setting parent context in submission payload
	//  - Fresh background context with configurable timeout (default 10m)
	//  - Polling for completion
	SpawnSubSaga(
		ctx context.Context,
		sagaTemplateId string,
		sagaInput map[string]string,
		originIdempotencyKey string,
		opts ...SubSagaOption,
	) (*SubSagaResult, error)
}

// SubSagaOption configures sub-saga spawning behavior
type SubSagaOption func(*subSagaOptions)

type subSagaOptions struct {
	timeout      time.Duration
	pollInterval time.Duration
}

// WithSubSagaTimeout sets the maximum wait time for the sub-saga to complete
func WithSubSagaTimeout(d time.Duration) SubSagaOption {
	return func(o *subSagaOptions) { o.timeout = d }
}

// WithSubSagaPollInterval sets the polling interval for checking sub-saga status
func WithSubSagaPollInterval(d time.Duration) SubSagaOption {
	return func(o *subSagaOptions) { o.pollInterval = d }
}

type sagaContextKey struct{}

// WithSagaContext stores a SagaContext in the given context
func WithSagaContext(ctx context.Context, sc SagaContext) context.Context {
	return context.WithValue(ctx, sagaContextKey{}, sc)
}

// GetSagaContext extracts the SagaContext from the given context.
// Returns nil if no SagaContext was set (e.g. executor not configured for sub-sagas).
func GetSagaContext(ctx context.Context) SagaContext {
	sc, _ := ctx.Value(sagaContextKey{}).(SagaContext)
	return sc
}

type defaultSagaContext struct {
	parentSagaInstanceId     string
	parentSagaStepInstanceId string
	rootSagaInstanceId       string
	sagaDepth                int
	clusterId                string
	sagaSubmitter            SagaSubmitter
	traxCtrlURL              string
}

func (sc *defaultSagaContext) ParentSagaInstanceId() string     { return sc.parentSagaInstanceId }
func (sc *defaultSagaContext) ParentSagaStepInstanceId() string { return sc.parentSagaStepInstanceId }
func (sc *defaultSagaContext) RootSagaInstanceId() string       { return sc.rootSagaInstanceId }
func (sc *defaultSagaContext) SagaDepth() int                   { return sc.sagaDepth }
func (sc *defaultSagaContext) ClusterId() string                { return sc.clusterId }

func (sc *defaultSagaContext) SpawnSubSaga(
	ctx context.Context,
	sagaTemplateId string,
	sagaInput map[string]string,
	originIdempotencyKey string,
	opts ...SubSagaOption,
) (*SubSagaResult, error) {
	options := &subSagaOptions{
		timeout:      10 * time.Minute,
		pollInterval: 2 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	subExec := NewSubSagaExecutor(
		sc.sagaSubmitter,
		sc.traxCtrlURL,
		WithPollInterval(options.pollInterval),
		WithMaxWaitTime(options.timeout),
	)

	// Use fresh background context to avoid inheriting parent's deadline
	subCtx, cancel := context.WithTimeout(context.Background(), options.timeout)
	defer cancel()

	common.L.Info(fmt.Sprintf(
		"spawning sub-saga '%s' from parent saga '%s' step '%s' (depth=%d, root=%s)",
		sagaTemplateId,
		sc.parentSagaInstanceId,
		sc.parentSagaStepInstanceId,
		sc.sagaDepth,
		sc.rootSagaInstanceId,
	))

	return subExec.SpawnAndWaitWithParent(
		subCtx,
		sc.clusterId,
		sagaTemplateId,
		sagaInput,
		originIdempotencyKey,
		sc.parentSagaInstanceId,
		sc.parentSagaStepInstanceId,
		sc.rootSagaInstanceId,
		sc.sagaDepth,
	)
}
