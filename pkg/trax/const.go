package trax

import "errors"

// Sentinel errors for store lookups
var (
	ErrSagaInstanceNotFound     = errors.New("saga instance not found")
	ErrSagaStepInstanceNotFound = errors.New("saga step instance not found")
)

type SagaStateEnum string

// states covering one saga instance's life cycle
const (
	// when the saga instance is created. this is the initial in-memory state
	// before it is persisted.
	SagaStateEnum_Unknown SagaStateEnum = "UNKNOWN"
	// initial state. when the saga instance is being run by the coordinator. this means
	// that its steps are being executed and the saga is not yet in a final state.
	SagaStateEnum_Running SagaStateEnum = "SAGA_STATE_ENUM_RUNNING"
	// when the saga instance has completed its execution successfully. this means
	// that all steps have been executed successfully and the saga is in a final state.
	SagaStateEnum_Committed SagaStateEnum = "SAGA_STATE_ENUM_COMMITTED"
	// when the saga instance has been compensated. this means that all compensating
	// actions have been executed successfully and the saga is in a final state. this
	// means that the saga instance has been reverted to its previous state (because of
	// a failure in one of its steps).
	SagaStateEnum_Compensated SagaStateEnum = "SAGA_STATE_ENUM_COMPENSATED"
	// when one of the steps is blocked and cannot proceed without human intervention.
	SagaStateEnum_Blocked SagaStateEnum = "SAGA_STATE_ENUM_BLOCKED"
	// when the saga instance is in an invalid state meaning it cannot be processed further.
	// this can happen happen when the states of the saga instance and its steps indicate
	// invalid transitions within the saga or saga-step processing state machines.
	//
	// when setting this state, nothing must change in the steps.
	//
	// resolving from this state requires a manual intervention and study of the invalid
	// state transitions and reporting a bug if needed.
	SagaStateEnum_InvalidState SagaStateEnum = "SAGA_STATE_ENUM_INVALID_STATE"

	// when a parent saga's coordinator requests compensation of a committed child saga.
	// the child saga's coordinator picks up this state and begins backward compensation walk.
	SagaStateEnum_CompensationRequested SagaStateEnum = "SAGA_STATE_ENUM_COMPENSATION_REQUESTED"

	// vvv STATES NOT BEING HANDLED FOR THE MOMENT vvv

	// when the saga instance is being paused by the coordinator.
	SagaStateEnum_Paused SagaStateEnum = "SAGA_STATE_ENUM_PAUSED"
	// when the saga instance is being cancelled by the coordinator for some
	// reason (user initiated, etc).
	SagaStateEnum_Cancelled SagaStateEnum = "SAGA_STATE_ENUM_CANCELLED"
)

type SagaStepStateEnum string

// persistent states covering one saga step's life cycle
const (
	SagaStepStateEnum_Unknown SagaStepStateEnum = "UNKNOWN"

	SagaStepStateEnum_ExecutionPending   SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_PENDING"
	SagaStepStateEnum_ExecutionCandidate SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_CANDIDATE"
	SagaStepStateEnum_ExecutionRunning   SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_RUNNING"
	SagaStepStateEnum_ExecutionSucceeded SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_SUCCEEDED"
	SagaStepStateEnum_ExecutionDone      SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_DONE"
	SagaStepStateEnum_ExecutionFailed    SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_FAILED"
	// in blocked state: an external agent (usually a human) must intervene to resolve this state
	SagaStepStateEnum_ExecutionBlocked SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_BLOCKED"
	SagaStepStateEnum_ExecutionAborted SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_EXECUTION_ABORTED"

	SagaStepStateEnum_CompensationPending   SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_PENDING"
	SagaStepStateEnum_CompensationCandidate SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_CANDIDATE"
	SagaStepStateEnum_CompensationRunning   SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_RUNNING"
	SagaStepStateEnum_CompensationSucceeded SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_SUCCEEDED"
	SagaStepStateEnum_CompensationDone      SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_DONE"
	SagaStepStateEnum_CompensationFailed    SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_FAILED"
	// in blocked state: an external agent (usually a human) must intervene to resolve this state
	SagaStepStateEnum_CompensationBlocked SagaStepStateEnum = "SAGA_STEP_STATE_ENUM_COMPENSATION_BLOCKED"
)

type ExecutionResultStatusEnum string

const (
	ExecutionResultStatusEnum_Unknown     ExecutionResultStatusEnum = "UNKNOWN"
	ExecutionResultStatusEnum_Success     ExecutionResultStatusEnum = "EXECUTION_RESULT_STATUS_ENUM_SUCCESS"
	ExecutionResultStatusEnum_InExecution ExecutionResultStatusEnum = "EXECUTION_RESULT_STATUS_ENUM_IN_EXECUTION"
	ExecutionResultStatusEnum_Failed      ExecutionResultStatusEnum = "EXECUTION_RESULT_STATUS_ENUM_FAILED"
	ExecutionResultStatusEnum_Retry       ExecutionResultStatusEnum = "EXECUTION_RESULT_STATUS_ENUM_RETRY"
	ExecutionResultStatusEnum_Error       ExecutionResultStatusEnum = "EXECUTION_RESULT_STATUS_ENUM_ERROR"
)
