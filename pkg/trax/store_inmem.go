package trax

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/xshyft/trax/pkg/common"
)

type inMemoryStore struct {
	sagaTemplates     map[string]*SagaTemplate
	sagaStepTemplates map[string]*SagaStepTemplate
	sagaInstances     map[string]*SagaInstance
	sagaStepInstances map[string]*SagaStepInstance
	clusters          map[string]*Cluster
	// annexes keyed by iid; cluster + saga lookups walk the map
	// since this is a test-only store and the volume is tiny.
	sagaAnnexes map[string]*SagaAnnex
}

func NewInMemoryStore() Store {
	return &inMemoryStore{
		sagaTemplates:     make(map[string]*SagaTemplate),
		sagaStepTemplates: make(map[string]*SagaStepTemplate),
		sagaInstances:     make(map[string]*SagaInstance),
		sagaStepInstances: make(map[string]*SagaStepInstance),
		clusters:          make(map[string]*Cluster),
		sagaAnnexes:       make(map[string]*SagaAnnex),
	}
}

func (s *inMemoryStore) Init(ctx context.Context) error {
	// No initialization needed for in-memory store
	return nil
}

func (s *inMemoryStore) Close() error {
	// No cleanup needed for in-memory store
	return nil
}

func (s *inMemoryStore) HealthCheck(ctx context.Context) error {
	// In-memory store is always healthy
	return nil
}

// LISTEN/NOTIFY methods - not supported for in-memory store
func (s *inMemoryStore) Listen(ctx context.Context, channel string) error {
	common.L.Warn("LISTEN/NOTIFY is not supported for in-memory store - falling back to polling only")
	return fmt.Errorf("LISTEN/NOTIFY not supported for in-memory store")
}

func (s *inMemoryStore) Unlisten(ctx context.Context, channel string) error {
	common.L.Warn("UNLISTEN is not supported for in-memory store")
	return fmt.Errorf("UNLISTEN not supported for in-memory store")
}

func (s *inMemoryStore) Notifications() <-chan *StoreNotification {
	// Return nil channel since notifications are not supported
	return nil
}

func (s *inMemoryStore) Notify(ctx context.Context, channel string, payload string) error {
	// No-op for in-memory store
	return nil
}

func (s *inMemoryStore) BeginTransaction(ctx context.Context) error {
	return nil
}

func (s *inMemoryStore) CommitTransaction(ctx context.Context) error {
	return nil
}

func (s *inMemoryStore) RollbackTransaction(ctx context.Context) error {
	return nil
}

func (s *inMemoryStore) SaveSagaInstanceIdempotently(ctx context.Context, saga *SagaInstance) (bool, error) {
	sagaIdempotencyKey := saga.SagaIdempotencyKey()
	if _, exists := s.sagaInstances[sagaIdempotencyKey]; exists {
		return false, nil
	}
	s.sagaInstances[sagaIdempotencyKey] = saga
	return true, nil
}

func (s *inMemoryStore) UpdateSagaState(ctx context.Context, saga *SagaInstance, state SagaStateEnum) error {
	sagaIdempotencyKey := saga.SagaIdempotencyKey()
	if _, exists := s.sagaInstances[sagaIdempotencyKey]; !exists {
		return ErrSagaInstanceNotFound
	}
	saga.State = state
	return nil
}

func (s *inMemoryStore) SaveSagaTemplateIdempotently(ctx context.Context, sagaTemplate *SagaTemplate) (bool, error) {
	if _, exists := s.sagaTemplates[sagaTemplate.TemplateId]; exists {
		return false, nil
	}
	s.sagaTemplates[sagaTemplate.TemplateId] = sagaTemplate
	return true, nil
}

func (s *inMemoryStore) GetSagaTemplate(ctx context.Context, templateId string) (*SagaTemplate, error) {
	template, exists := s.sagaTemplates[templateId]
	if !exists {
		return nil, fmt.Errorf("saga template not found")
	}
	return template, nil
}

func (s *inMemoryStore) SaveSagaStepTemplateIdempotently(ctx context.Context, sagaStepTemplate *SagaStepTemplate) (bool, error) {
	if _, exists := s.sagaStepTemplates[sagaStepTemplate.TemplateId]; exists {
		return false, nil
	}
	s.sagaStepTemplates[sagaStepTemplate.TemplateId] = sagaStepTemplate
	return true, nil
}

func (s *inMemoryStore) GetSagaStepTemplate(ctx context.Context, templateId string) (*SagaStepTemplate, error) {
	template, exists := s.sagaStepTemplates[templateId]
	if !exists {
		return nil, fmt.Errorf("saga step template not found")
	}
	return template, nil
}

func (s *inMemoryStore) ListSagaTemplates(ctx context.Context) ([]*SagaTemplate, error) {
	var result []*SagaTemplate
	for _, template := range s.sagaTemplates {
		result = append(result, template)
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaTemplateIds(ctx context.Context) ([]string, error) {
	var result []string
	for templateId := range s.sagaTemplates {
		result = append(result, templateId)
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaStepTemplates(ctx context.Context) ([]*SagaStepTemplate, error) {
	var result []*SagaStepTemplate
	for _, template := range s.sagaStepTemplates {
		result = append(result, template)
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaStepTemplateIds(ctx context.Context) ([]string, error) {
	var result []string
	for templateId := range s.sagaStepTemplates {
		result = append(result, templateId)
	}
	return result, nil
}

func (s *inMemoryStore) UpdateSagaTemplate(ctx context.Context, sagaTemplate *SagaTemplate) error {
	if _, exists := s.sagaTemplates[sagaTemplate.TemplateId]; !exists {
		return fmt.Errorf("saga template not found: %s", sagaTemplate.TemplateId)
	}
	s.sagaTemplates[sagaTemplate.TemplateId] = sagaTemplate
	return nil
}

func (s *inMemoryStore) DeleteSagaTemplate(ctx context.Context, templateId string) error {
	if _, exists := s.sagaTemplates[templateId]; !exists {
		return fmt.Errorf("saga template not found: %s", templateId)
	}
	// Delete associated step templates first
	for id, step := range s.sagaStepTemplates {
		if step.SagaTemplateId == templateId {
			delete(s.sagaStepTemplates, id)
		}
	}
	delete(s.sagaTemplates, templateId)
	return nil
}

func (s *inMemoryStore) UpdateSagaStepTemplate(ctx context.Context, sagaStepTemplate *SagaStepTemplate) error {
	if _, exists := s.sagaStepTemplates[sagaStepTemplate.TemplateId]; !exists {
		return fmt.Errorf("saga step template not found: %s", sagaStepTemplate.TemplateId)
	}
	s.sagaStepTemplates[sagaStepTemplate.TemplateId] = sagaStepTemplate
	return nil
}

func (s *inMemoryStore) DeleteSagaStepTemplate(ctx context.Context, templateId string) error {
	if _, exists := s.sagaStepTemplates[templateId]; !exists {
		return fmt.Errorf("saga step template not found: %s", templateId)
	}
	delete(s.sagaStepTemplates, templateId)
	return nil
}

func (s *inMemoryStore) SaveSagaStepInstanceIdempotently(ctx context.Context, sagaStepInstance *SagaStepInstance) (bool, error) {
	sagaStepIdempotencyKey := sagaStepInstance.SagaIdempotencyKey()
	if _, exists := s.sagaStepInstances[sagaStepIdempotencyKey]; exists {
		return false, nil
	}
	s.sagaStepInstances[sagaStepIdempotencyKey] = sagaStepInstance
	return true, nil
}

func (s *inMemoryStore) UpdateSagaStepState(ctx context.Context, sagaStep *SagaStepInstance, state SagaStepStateEnum) error {
	sagaStepIdempotencyKey := sagaStep.SagaIdempotencyKey()
	if _, exists := s.sagaStepInstances[sagaStepIdempotencyKey]; !exists {
		return ErrSagaStepInstanceNotFound
	}
	sagaStep.State = state
	return nil
}

func (s *inMemoryStore) UpdateSagaStepResult(ctx context.Context, sagaStepInstance *SagaStepInstance) error {
	sagaStepIdempotencyKey := sagaStepInstance.SagaIdempotencyKey()
	if _, exists := s.sagaStepInstances[sagaStepIdempotencyKey]; !exists {
		return ErrSagaStepInstanceNotFound
	}
	// For in-memory store, the saga step instance is already updated in place
	// since we're working with the same object reference
	return nil
}

func (s *inMemoryStore) UpdateSagaStepCompensationResult(ctx context.Context, sagaStepInstance *SagaStepInstance) error {
	sagaStepIdempotencyKey := sagaStepInstance.SagaIdempotencyKey()
	if _, exists := s.sagaStepInstances[sagaStepIdempotencyKey]; !exists {
		return ErrSagaStepInstanceNotFound
	}
	// For in-memory store, the saga step instance is already updated in place
	// since we're working with the same object reference
	return nil
}

func (s *inMemoryStore) UpdateSagaStepInstanceExecutionHistory(ctx context.Context, sagaStepInstance *SagaStepInstance) error {
	sagaStepIdempotencyKey := sagaStepInstance.SagaIdempotencyKey()
	if _, exists := s.sagaStepInstances[sagaStepIdempotencyKey]; !exists {
		return ErrSagaStepInstanceNotFound
	}
	// For in-memory store, the saga step instance is already updated in place
	// since we're working with the same object reference
	return nil
}

func (s *inMemoryStore) GetSagaInstance(ctx context.Context, clusterId, instanceId string) (*SagaInstance, error) {
	for _, saga := range s.sagaInstances {
		if saga.InstanceId == instanceId {
			return saga, nil
		}
	}
	return nil, ErrSagaInstanceNotFound
}

func (s *inMemoryStore) GetSagaStepInstance(ctx context.Context, clusterId, instanceId string) (*SagaStepInstance, error) {
	for _, sagaStep := range s.sagaStepInstances {
		if sagaStep.ClusterId == clusterId && sagaStep.InstanceId == instanceId {
			return sagaStep, nil
		}
	}
	return nil, ErrSagaStepInstanceNotFound
}

func (s *inMemoryStore) GetSagaStepInstancesByAffinityAndOneOfSagaStatesAndOneOfSagaStepStates(
	ctx context.Context,
	clusterId, affinity string,
	sagaStates []SagaStateEnum,
	sagaStepStates []SagaStepStateEnum,
) ([]*SagaStepInstance, error) {
	var result []*SagaStepInstance

	// Look through all saga step instances to find matching affinity and states
	for _, sagaStep := range s.sagaStepInstances {
		if sagaStep.Affinity != affinity {
			continue
		}

		// Find the corresponding saga instance to check its state
		var sagaInstance *SagaInstance
		for _, saga := range s.sagaInstances {
			if saga.InstanceId == sagaStep.SagaInstanceId {
				sagaInstance = saga
				break
			}
		}
		if sagaInstance == nil {
			continue // Skip if saga instance not found
		}

		// Check if saga state matches any of the requested saga states
		sagaStateMatches := false
		for _, state := range sagaStates {
			if sagaInstance.State == state {
				sagaStateMatches = true
				break
			}
		}
		if !sagaStateMatches {
			continue
		}

		// Check if step state matches any of the requested saga step states
		stepStateMatches := false
		for _, state := range sagaStepStates {
			if sagaStep.State == state {
				stepStateMatches = true
				break
			}
		}
		if !stepStateMatches {
			continue
		}

		result = append(result, sagaStep)
	}
	return result, nil
}

func (s *inMemoryStore) GetSagaStepBySagaIdempotencyKey(ctx context.Context, clusterId, sagaStepIdempotentId string) (*SagaStepInstance, error) {
	sagaStep, exists := s.sagaStepInstances[sagaStepIdempotentId]
	if !exists {
		return nil, ErrSagaStepInstanceNotFound
	}
	return sagaStep, nil
}

// List methods implementation for inMemoryStore
func (s *inMemoryStore) ListSagaInstances(ctx context.Context, clusterId string) ([]*SagaInstance, error) {
	var result []*SagaInstance
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaInstancesPaginated(ctx context.Context, clusterId string, opts *common.QueryOptions) ([]*SagaInstance, int, error) {
	if opts == nil {
		opts = &common.QueryOptions{}
	}
	needle := strings.ToLower(strings.TrimSpace(opts.Search))
	stateFilter := opts.Filters["state"]
	templateFilter := opts.Filters["saga_template_id"]
	submitterFilter := opts.Filters["saga_submitter_id"]
	var all []*SagaInstance
	for _, instance := range s.sagaInstances {
		if instance.ClusterId != clusterId {
			continue
		}
		if stateFilter != "" && string(instance.State) != stateFilter {
			continue
		}
		if templateFilter != "" && instance.SagaTemplateId != templateFilter {
			continue
		}
		if submitterFilter != "" && instance.SagaSubmitterId != submitterFilter {
			continue
		}
		if needle != "" {
			// Mirror the psqlStore search whitelist as faithfully as the
			// in-memory representation allows — every visible text field
			// (JSONB columns marshalled to lowercase JSON for the haystack).
			labelsJSON, _ := json.Marshal(instance.Labels)
			tagsJSON, _ := json.Marshal(instance.Tags)
			metadataJSON, _ := json.Marshal(instance.Metadata)
			inputJSON, _ := json.Marshal(instance.Input)
			hay := strings.ToLower(
				instance.InstanceId + " " +
					instance.ZoneId + " " +
					instance.TraceId + " " +
					instance.ExecutionId + " " +
					instance.SagaSubmitterId + " " +
					instance.Origin + " " +
					instance.OriginIdempotencyKey + " " +
					string(labelsJSON) + " " +
					string(tagsJSON) + " " +
					string(metadataJSON) + " " +
					string(inputJSON) + " " +
					string(instance.State) + " " +
					instance.SagaTemplateId + " " +
					instance.ParentSagaInstanceId + " " +
					instance.ParentSagaStepInstanceId + " " +
					instance.RootSagaInstanceId + " " +
					instance.CompensationReason,
			)
			if !strings.Contains(hay, needle) {
				continue
			}
		}
		all = append(all, instance)
	}

	// Sorting — in-memory stub honours the primary sort column only;
	// production reads come from psqlStore which handles multi-column.
	if len(opts.SortBy) > 0 {
		primary := opts.SortBy[0]
		desc := strings.EqualFold(primary.Direction, "DESC")
		sort.SliceStable(all, func(i, j int) bool {
			a, b := inMemorySagaSortKey(all[i], primary.Column), inMemorySagaSortKey(all[j], primary.Column)
			if desc {
				return a > b
			}
			return a < b
		})
	}

	totalCount := len(all)

	offset := opts.Offset
	limit := opts.Limit
	if offset < 0 {
		offset = 0
	}
	if offset >= totalCount {
		return []*SagaInstance{}, totalCount, nil
	}

	end := totalCount
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}

	return all[offset:end], totalCount, nil
}

// inMemorySagaSortKey returns a comparable string for the named column. Used
// only by the in-memory stub; columns it doesn't recognize fall back to ""
// (stable, but undefined) — psqlStore is the production path.
func inMemorySagaSortKey(s *SagaInstance, column string) string {
	switch column {
	case "instance_id":
		return s.InstanceId
	case "created_at":
		return strconv.FormatInt(s.CreatedAt, 10)
	case "updated_at":
		return strconv.FormatInt(s.UpdatedAt, 10)
	case "state":
		return string(s.State)
	case "saga_template_id":
		return s.SagaTemplateId
	case "saga_submitter_id":
		return s.SagaSubmitterId
	case "trace_id":
		return s.TraceId
	case "execution_id":
		return s.ExecutionId
	case "zone_id":
		return s.ZoneId
	default:
		return ""
	}
}

func (s *inMemoryStore) ListSagaInstanceIds(ctx context.Context, clusterId string) ([]string, error) {
	var result []string
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId {
			result = append(result, instance.InstanceId)
		}
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaStepInstances(ctx context.Context, clusterId string) ([]*SagaStepInstance, error) {
	var result []*SagaStepInstance
	for _, instance := range s.sagaStepInstances {
		if instance.ClusterId == clusterId {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaStepInstanceIds(ctx context.Context, clusterId string) ([]string, error) {
	var result []string
	for _, instance := range s.sagaStepInstances {
		if instance.ClusterId == clusterId {
			result = append(result, instance.InstanceId)
		}
	}
	return result, nil
}

func (s *inMemoryStore) ListSagaStepInstancesBySagaInstanceId(ctx context.Context, clusterId, sagaInstanceId string) ([]*SagaStepInstance, error) {
	var result []*SagaStepInstance
	for _, instance := range s.sagaStepInstances {
		if instance.ClusterId == clusterId && instance.SagaInstanceId == sagaInstanceId {
			result = append(result, instance)
		}
	}
	return result, nil
}

// Sub-saga hierarchy queries for inMemoryStore
func (s *inMemoryStore) GetChildSagaInstances(ctx context.Context, clusterId, parentSagaInstanceId string) ([]*SagaInstance, error) {
	var result []*SagaInstance
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId && instance.ParentSagaInstanceId == parentSagaInstanceId {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (s *inMemoryStore) GetSagaHierarchy(ctx context.Context, clusterId, rootSagaInstanceId string) ([]*SagaInstance, error) {
	var result []*SagaInstance
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId && instance.RootSagaInstanceId == rootSagaInstanceId {
			result = append(result, instance)
		}
	}
	return result, nil
}

func (s *inMemoryStore) TriggerSagaCompensation(ctx context.Context, clusterId, sagaInstanceId string) error {
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId && instance.InstanceId == sagaInstanceId {
			if instance.State != SagaStateEnum_Committed {
				return fmt.Errorf("saga instance %s is not in COMMITTED state (current: %s)", sagaInstanceId, instance.State)
			}
			instance.State = SagaStateEnum_CompensationRequested
			return nil
		}
	}
	return ErrSagaInstanceNotFound
}

func (s *inMemoryStore) ForceMarkSagaCompensated(ctx context.Context, clusterId, sagaInstanceId, reason string) error {
	if reason == "" {
		return fmt.Errorf("reason is required for force-mark compensated")
	}
	for _, instance := range s.sagaInstances {
		if instance.ClusterId == clusterId && instance.InstanceId == sagaInstanceId {
			if instance.State != SagaStateEnum_Blocked {
				return fmt.Errorf("saga instance %s is not in BLOCKED state (current: %s)", sagaInstanceId, instance.State)
			}
			instance.State = SagaStateEnum_Compensated
			instance.CompensationReason = "[FORCE-MARKED] " + reason
			return nil
		}
	}
	return ErrSagaInstanceNotFound
}

// Cluster CRUD implementation for inMemoryStore
func (s *inMemoryStore) SaveClusterIdempotently(ctx context.Context, cluster *Cluster) (bool, error) {
	if _, exists := s.clusters[cluster.Id]; exists {
		return false, nil
	}
	s.clusters[cluster.Id] = cluster
	return true, nil
}

func (s *inMemoryStore) GetCluster(ctx context.Context, id string) (*Cluster, error) {
	cluster, exists := s.clusters[id]
	if !exists {
		return nil, fmt.Errorf("cluster not found")
	}
	return cluster, nil
}

func (s *inMemoryStore) UpdateCluster(ctx context.Context, cluster *Cluster) error {
	if _, exists := s.clusters[cluster.Id]; !exists {
		return fmt.Errorf("cluster not found")
	}
	s.clusters[cluster.Id] = cluster
	return nil
}

func (s *inMemoryStore) DeleteCluster(ctx context.Context, id string) error {
	if _, exists := s.clusters[id]; !exists {
		return fmt.Errorf("cluster not found")
	}
	delete(s.clusters, id)
	return nil
}

func (s *inMemoryStore) ListClusters(ctx context.Context) ([]*Cluster, error) {
	var result []*Cluster
	for _, cluster := range s.clusters {
		result = append(result, cluster)
	}
	return result, nil
}

func (s *inMemoryStore) ListClusterIds(ctx context.Context) ([]string, error) {
	var result []string
	for id := range s.clusters {
		result = append(result, id)
	}
	return result, nil
}

func (s *inMemoryStore) CreateSagaAnnex(ctx context.Context, annex *SagaAnnex) error {
	if annex == nil || annex.Iid == "" || annex.SagaInstanceId == "" {
		return fmt.Errorf("annex iid and saga_instance_id are required")
	}
	saga, ok := s.sagaInstances[annex.SagaInstanceId]
	if !ok || saga.ClusterId != annex.ClusterId {
		return fmt.Errorf("saga instance not found: %s", annex.SagaInstanceId)
	}
	cp := *annex
	if cp.ContentLength == 0 {
		cp.ContentLength = int64(len(cp.ContentData))
	}
	s.sagaAnnexes[cp.Iid] = &cp
	saga.AnnexIids = append(saga.AnnexIids, cp.Iid)
	return nil
}

func (s *inMemoryStore) ListSagaAnnexes(ctx context.Context, clusterId, sagaInstanceId string) ([]*SagaAnnex, error) {
	var result []*SagaAnnex
	for _, a := range s.sagaAnnexes {
		if a.ClusterId == clusterId && a.SagaInstanceId == sagaInstanceId {
			meta := *a
			meta.ContentData = nil
			result = append(result, &meta)
		}
	}
	return result, nil
}

func (s *inMemoryStore) GetSagaAnnexBytes(ctx context.Context, clusterId, annexIid string) (*SagaAnnex, error) {
	a, ok := s.sagaAnnexes[annexIid]
	if !ok || a.ClusterId != clusterId {
		return nil, fmt.Errorf("annex not found: %s", annexIid)
	}
	cp := *a
	return &cp, nil
}
