package trax

import "context"

type SagaIdempotencyKeyStatusEnum string

const (
	SagaIdempotencyKeyStatusEnum_Unknown    SagaIdempotencyKeyStatusEnum = "UNKNOWN"
	SagaIdempotencyKeyStatusEnum_NotSeen    SagaIdempotencyKeyStatusEnum = "IDEMPOTENT_KEY_STATUS_ENUM_NOT_SEEN"
	SagaIdempotencyKeyStatusEnum_InProgress SagaIdempotencyKeyStatusEnum = "IDEMPOTENT_KEY_STATUS_ENUM_IN_PROGRESS"
	SagaIdempotencyKeyStatusEnum_Completed  SagaIdempotencyKeyStatusEnum = "IDEMPOTENT_KEY_STATUS_ENUM_COMPLETED"
)

type IdempotentServiceExecutionResult struct {
	Result map[string]string
	Error  error
}

// IdempotentService provides methods to handle idempotent operations.
type IdempotentService interface {
	GetIdempotentKeyExecutionStatus(
		ctx context.Context,
		sagaIdempotencyKey string,
	) (SagaIdempotencyKeyStatusEnum, error)
	ExecuteSync(
		ctx context.Context,
		sagaIdempotencyKey string,
		input map[string]string,
	) (*IdempotentServiceExecutionResult, error)
	ExecuteAsync(
		ctx context.Context,
		sagaIdempotencyKey string,
		input map[string]string,
		cb func(*IdempotentServiceExecutionResult, error),
	)
	GetIdempotentKeyCompensationStatus(
		ctx context.Context,
		sagaIdempotencyKey string,
	) (SagaIdempotencyKeyStatusEnum, error)
	CompensateSync(
		ctx context.Context,
		sagaIdempotencyKey string,
		input map[string]string,
	) (*IdempotentServiceExecutionResult, error)
	CompensateAsync(
		ctx context.Context,
		sagaIdempotencyKey string,
		input map[string]string,
		cb func(*IdempotentServiceExecutionResult, error),
	)
}
