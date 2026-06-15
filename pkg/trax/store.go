package trax

import (
	"context"

	"github.com/kamcpp/trax/pkg/common"
)

// StoreNotification represents a notification from the database
type StoreNotification struct {
	Channel string
	Payload string
}

type Store interface {
	Init(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error
	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error

	// LISTEN/NOTIFY support for event-driven saga processing
	// Listen starts listening on a channel for notifications
	Listen(ctx context.Context, channel string) error
	// Unlisten stops listening on a channel
	Unlisten(ctx context.Context, channel string) error
	// Notifications returns a channel for receiving database notifications
	// Returns nil if the store implementation doesn't support notifications
	Notifications() <-chan *StoreNotification
	// Notify sends a notification to a channel (used when inserting saga steps)
	Notify(ctx context.Context, channel string, payload string) error

	SaveSagaTemplateIdempotently(ctx context.Context, sagaTemplate *SagaTemplate) (bool, error)
	GetSagaTemplate(ctx context.Context, id string) (*SagaTemplate, error)
	ListSagaTemplates(ctx context.Context) ([]*SagaTemplate, error)
	ListSagaTemplateIds(ctx context.Context) ([]string, error)
	UpdateSagaTemplate(ctx context.Context, sagaTemplate *SagaTemplate) error
	DeleteSagaTemplate(ctx context.Context, templateId string) error

	SaveSagaStepTemplateIdempotently(ctx context.Context, sagaStepTemplate *SagaStepTemplate) (bool, error)
	GetSagaStepTemplate(ctx context.Context, id string) (*SagaStepTemplate, error)
	ListSagaStepTemplates(ctx context.Context) ([]*SagaStepTemplate, error)
	ListSagaStepTemplateIds(ctx context.Context) ([]string, error)
	UpdateSagaStepTemplate(ctx context.Context, sagaStepTemplate *SagaStepTemplate) error
	DeleteSagaStepTemplate(ctx context.Context, templateId string) error

	SaveSagaInstanceIdempotently(ctx context.Context, sagaInstance *SagaInstance) (bool, error)
	UpdateSagaState(ctx context.Context, sagaInstance *SagaInstance, state SagaStateEnum) error
	GetSagaInstance(ctx context.Context, clusterId, id string) (*SagaInstance, error)

	SaveSagaStepInstanceIdempotently(ctx context.Context, sagaStepInstance *SagaStepInstance) (bool, error)
	UpdateSagaStepState(ctx context.Context, sagaStepInstance *SagaStepInstance, state SagaStepStateEnum) error
	UpdateSagaStepResult(ctx context.Context, sagaStepInstance *SagaStepInstance) error
	UpdateSagaStepCompensationResult(ctx context.Context, sagaStepInstance *SagaStepInstance) error
	UpdateSagaStepInstanceExecutionHistory(ctx context.Context, sagaStepInstance *SagaStepInstance) error
	GetSagaStepInstance(ctx context.Context, clusterId, id string) (*SagaStepInstance, error)
	GetSagaStepBySagaIdempotencyKey(ctx context.Context, clusterId, sagaStepIdempotentId string) (*SagaStepInstance, error)
	GetSagaStepInstancesByAffinityAndOneOfSagaStatesAndOneOfSagaStepStates(
		ctx context.Context,
		clusterId, affinity string,
		sagaStates []SagaStateEnum,
		sagaStepStates []SagaStepStateEnum,
	) ([]*SagaStepInstance, error)

	// List methods for saga instances
	ListSagaInstances(ctx context.Context, clusterId string) ([]*SagaInstance, error)
	// ListSagaInstancesPaginated returns a paginated, sorted, optionally
	// search-filtered slice. opts.Search applies case-insensitive ILIKE
	// across every visible text + JSONB column on saga_instances.
	// opts.SortBy may carry one or more columns (multi-column ORDER BY);
	// the default — when SortBy is empty — is `created_at DESC, instance_id ASC`.
	// Total count reflects the filtered set.
	ListSagaInstancesPaginated(ctx context.Context, clusterId string, opts *common.QueryOptions) ([]*SagaInstance, int, error)
	ListSagaInstanceIds(ctx context.Context, clusterId string) ([]string, error)

	// List methods for saga step instances
	ListSagaStepInstances(ctx context.Context, clusterId string) ([]*SagaStepInstance, error)
	ListSagaStepInstanceIds(ctx context.Context, clusterId string) ([]string, error)
	ListSagaStepInstancesBySagaInstanceId(ctx context.Context, clusterId, sagaInstanceId string) ([]*SagaStepInstance, error)

	// Sub-saga hierarchy queries
	// GetChildSagaInstances returns all direct children of a saga instance
	GetChildSagaInstances(ctx context.Context, clusterId, parentSagaInstanceId string) ([]*SagaInstance, error)
	// GetSagaHierarchy returns all saga instances sharing the same root saga instance ID
	GetSagaHierarchy(ctx context.Context, clusterId, rootSagaInstanceId string) ([]*SagaInstance, error)
	// TriggerSagaCompensation sets a committed saga to COMPENSATION_REQUESTED state,
	// which the coordinator will pick up and begin compensating.
	TriggerSagaCompensation(ctx context.Context, clusterId, sagaInstanceId string) error

	// ForceMarkSagaCompensated is an operator-override that flips a
	// BLOCKED saga directly to COMPENSATED (terminal). Used when a
	// compensation step has wedged the saga in a state it can't
	// recover from on its own — e.g. trying to delete a row that's
	// already gone — and the operator has confirmed via inspection
	// that the saga's effects have been (or never were) applied.
	// The [reason] is required and lands on the saga's
	// `compensation_reason` column for audit. Returns an error when
	// the saga isn't currently BLOCKED so the override can't
	// accidentally short-circuit a healthy saga.
	ForceMarkSagaCompensated(ctx context.Context, clusterId, sagaInstanceId, reason string) error

	// Saga annex storage. Trax is the owner of binary attachments
	// tied to a saga; gateways (csdmsggw, …) push bytes here after
	// the saga is created and read them back via traxctrl.
	//
	// CreateSagaAnnex writes a single annex row and appends its iid
	// to the saga's `annex_iids` array. Returns ErrSagaInstanceNotFound
	// when the parent saga doesn't exist (annexes are saga-owned —
	// orphaned bytes aren't allowed).
	CreateSagaAnnex(ctx context.Context, annex *SagaAnnex) error

	// ListSagaAnnexes returns the metadata (no bytes) for every
	// annex attached to the given saga, ordered oldest-first.
	ListSagaAnnexes(ctx context.Context, clusterId, sagaInstanceId string) ([]*SagaAnnex, error)

	// GetSagaAnnexBytes returns a single annex including its raw
	// content bytes. Caller is responsible for the saga ↔ annex
	// consistency check (don't allow cross-saga retrieval).
	GetSagaAnnexBytes(ctx context.Context, clusterId, annexIid string) (*SagaAnnex, error)

	// Cluster CRUD operations
	SaveClusterIdempotently(ctx context.Context, cluster *Cluster) (bool, error)
	GetCluster(ctx context.Context, id string) (*Cluster, error)
	UpdateCluster(ctx context.Context, cluster *Cluster) error
	DeleteCluster(ctx context.Context, id string) error
	ListClusters(ctx context.Context) ([]*Cluster, error)
	ListClusterIds(ctx context.Context) ([]string, error)
}
