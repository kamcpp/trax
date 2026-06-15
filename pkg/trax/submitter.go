package trax

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/execpl"
)

type PostAnnounceSagaSubmitterRequest struct {
	SagaSubmitterId string `json:"saga_submitter_id" binding:"required"`
}

type SubmitterNodeNames struct {
	Inbox  string
	Outbox string
}

type PostAnnounceSagaSubmitterResponse struct {
	ClusterIds             []string                       `json:"cluster_ids"`
	NodeNamesPerClusterMap map[string]*SubmitterNodeNames `json:"node_names_per_cluster"`
}

// SagaCompletionResult contains the result of waiting for a saga to complete
type SagaCompletionResult struct {
	// SagaInstanceId is the ID of the saga instance
	SagaInstanceId string

	// State is the final state of the saga
	State SagaStateEnum

	// Outputs contains the combined outputs from all successful steps
	Outputs map[string]string

	// Error contains any error message if the saga failed
	Error string
}

type SagaSubmitter interface {
	Id() string
	StartAnnouncement(ctx context.Context)
	IsReadyToAcceptSagaSubmissionRequests() bool
	// IsReadyWithClusters checks if the submitter is ready AND has at least one cluster ID.
	// This is stricter than IsReadyToAcceptSagaSubmissionRequests which only checks the ready flag.
	IsReadyWithClusters() bool
	WaitUntilReadyToAcceptSagaSubmissionRequests(ctx context.Context) error
	// GetDefaultClusterId returns the first available cluster ID.
	// TODO(kam): This is a temporary solution. Each participant should be mapped to their own cluster.
	// See TODO.md for tracking this work.
	GetDefaultClusterId() string
	// GetClusterIds returns all available cluster IDs.
	GetClusterIds() []string
	// GetTraxCtrlURL returns the traxctrl service base URL
	GetTraxCtrlURL() string
	// SetTraxCtrlURL sets the traxctrl service base URL (must be called before using WaitForSagaCompletion)
	SetTraxCtrlURL(url string)
	SubmitSaga(
		ctx context.Context,
		participantId string,
		traceId string,
		zoneId string,
		origin string,
		originIdempotencyKey string,
		issuer string,
		referrer string,
		tags []string,
		metadata map[string]string,
		sagaTemplateId string,
		sagaInput map[string]string,
	) (sagaInstanceId string, err error)
	// SubmitSubSaga submits a saga with parent context for automatic parent-child registration.
	// The parent context fields are included in the SagaSubmissionRequestPayload so the
	// coordinator stores the parent-child relationship when creating the saga instance.
	SubmitSubSaga(
		ctx context.Context,
		participantId string,
		traceId string,
		zoneId string,
		origin string,
		originIdempotencyKey string,
		issuer string,
		referrer string,
		tags []string,
		metadata map[string]string,
		sagaTemplateId string,
		sagaInput map[string]string,
		parentSagaInstanceId string,
		parentSagaStepInstanceId string,
		rootSagaInstanceId string,
		sagaDepth int,
	) (sagaInstanceId string, err error)
	// WaitForSagaCompletion polls the traxctrl API until the saga reaches a terminal state (Committed, Compensated, or Blocked)
	WaitForSagaCompletion(
		ctx context.Context,
		clusterId string,
		sagaInstanceId string,
		pollInterval time.Duration,
		maxWaitTime time.Duration,
	) (*SagaCompletionResult, error)
	// ResetForTesting clears the cached cluster state and forces re-announcement.
	// This is used during E2E testing when the database is switched dynamically.
	// After calling this method, the submitter will re-announce on its next interval
	// and pick up the new cluster IDs from the switched database.
	ResetForTesting()
}

type defaultSagaSubmitter struct {
	id       string
	mqClient MQClient

	mu                                  sync.RWMutex
	clusterIds                          []string
	nodeNames                           map[string]*SubmitterNodeNames
	readyToAcceptSagaSubmissionRequests bool
	traxCtrlURL                         string
	httpClient                          *http.Client
	// consumersStarted tracks which inbox nodes already have consumers to prevent
	// creating duplicate consumers on re-announcement (which causes memory leaks)
	consumersStarted map[string]bool
}

func NewDefaultSagaSubmitter(id string, mqClient MQClient) SagaSubmitter {
	return &defaultSagaSubmitter{
		id:       id,
		mqClient: mqClient,

		clusterIds:       []string{},
		nodeNames:        make(map[string]*SubmitterNodeNames),
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		consumersStarted: make(map[string]bool),
	}
}

func (s *defaultSagaSubmitter) Id() string {
	return s.id
}

func (s *defaultSagaSubmitter) IsReadyToAcceptSagaSubmissionRequests() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readyToAcceptSagaSubmissionRequests
}

// IsReadyWithClusters checks if the submitter is ready AND has at least one cluster ID.
// This is stricter than IsReadyToAcceptSagaSubmissionRequests which only checks the ready flag.
// Use this in tests to ensure the submitter has actually received cluster IDs from coordinators.
func (s *defaultSagaSubmitter) IsReadyWithClusters() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readyToAcceptSagaSubmissionRequests && len(s.clusterIds) > 0
}

// GetDefaultClusterId returns the first available cluster ID.
// TODO(kam): This is a temporary solution. Each participant should be mapped to their own cluster.
func (s *defaultSagaSubmitter) GetDefaultClusterId() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.clusterIds) > 0 {
		return s.clusterIds[0]
	}
	return ""
}

// GetClusterIds returns a copy of all available cluster IDs.
func (s *defaultSagaSubmitter) GetClusterIds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, len(s.clusterIds))
	copy(result, s.clusterIds)
	return result
}

// GetTraxCtrlURL returns the traxctrl service base URL
func (s *defaultSagaSubmitter) GetTraxCtrlURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.traxCtrlURL
}

// SetTraxCtrlURL sets the traxctrl service base URL
func (s *defaultSagaSubmitter) SetTraxCtrlURL(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traxCtrlURL = url
}

func (s *defaultSagaSubmitter) WaitUntilReadyToAcceptSagaSubmissionRequests(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Use IsReadyWithClusters to ensure we have actual cluster IDs, not just "ready" flag.
			// This prevents the race condition where the submitter becomes "ready" with empty clusters
			// because the coordinator hasn't loaded cluster IDs from the new database yet.
			if s.IsReadyWithClusters() {
				clusterIds := s.GetClusterIds() // Thread-safe access
				common.L.Info(fmt.Sprintf("saga submitter '%s' is now ready to accept submissions with clusters: %v", s.id, clusterIds), common.F(ctx)...)
				return nil
			}
		}
	}
}

// Exponential backoff constants for announcement retries after failure.
// On transient coordinator unavailability, the submitter retries quickly (1s, 2s, 4s, ...)
// before falling back to the normal announcement interval.
const (
	announcementBackoffInitial    = 1 * time.Second
	announcementBackoffMultiplier = 2.0
	announcementBackoffMaxRetries = 5
)

// announceToCoordinator performs a single HTTP POST to the coordinator's announce endpoint.
func (s *defaultSagaSubmitter) announceToCoordinator(baseUrl string) (*http.Response, error) {
	postBody := PostAnnounceSagaSubmitterRequest{
		SagaSubmitterId: s.id,
	}
	postBodyBytes, err := json.Marshal(postBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post body: %w", err)
	}
	return http.Post(baseUrl+"/saga-submitter/announce", "application/json", bytes.NewBuffer(postBodyBytes))
}

// processAnnouncementResponse decodes a successful announcement response,
// sets the submitter as ready, and starts inbox consumers for new clusters.
func (s *defaultSagaSubmitter) processAnnouncementResponse(ctx context.Context, resp *http.Response) {
	var respBody PostAnnounceSagaSubmitterResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		common.L.Warn(fmt.Sprintf(
			"failed to decode response body for saga submitter '%s': %v", s.id, err), common.F(ctx)...)
		return
	}
	common.L.Debug(fmt.Sprintf(
		"successfully announced saga submitter '%s': resp: %+v",
		s.id, respBody), common.F(ctx)...)

	s.mu.Lock()
	if !s.readyToAcceptSagaSubmissionRequests {
		s.readyToAcceptSagaSubmissionRequests = true
		s.clusterIds = respBody.ClusterIds
		common.L.Info(fmt.Sprintf("[OK] set saga-submitter '%s' as READY with clusters: %v", s.id, s.clusterIds), common.F(ctx)...)
	}

	// Start consumers for inbox nodes that don't already have one.
	// This is done outside the readyToAcceptSagaSubmissionRequests check to handle
	// new clusters being added dynamically, while still preventing duplicate consumers.
	for _, clusterId := range respBody.ClusterIds {
		s.nodeNames[clusterId] = &SubmitterNodeNames{
			Inbox:  respBody.NodeNamesPerClusterMap[clusterId].Inbox,
			Outbox: respBody.NodeNamesPerClusterMap[clusterId].Outbox,
		}
		inboxNodeName := respBody.NodeNamesPerClusterMap[clusterId].Inbox

		// Only create consumer if one doesn't already exist for this inbox node.
		// This prevents memory leaks from duplicate consumers on re-announcement.
		if !s.consumersStarted[inboxNodeName] {
			s.consumersStarted[inboxNodeName] = true
			common.L.Info(fmt.Sprintf(
				"starting consumer for saga submitter '%s' on inbox node: %s",
				s.id, inboxNodeName), common.F(ctx)...)
			s.mqClient.ConsumeNodeAsync(ctx, inboxNodeName,
				func(ctx context.Context, messageType, contentType string, msg *TraxMessage) error {
					if len(msg.Payloads) == 0 {
						common.L.Warn(fmt.Sprintf(
							"received empty message for saga submitter '%s': %s", s.id, msg.Json()), common.F(ctx)...)
						// TODO(kam): maybe move to dead letter queue. for now, drop the message
						return nil
					}
					switch msg.Payloads[0].Type {
					case SagaPayloadType_SagaSubmissionSuccess:
						common.L.Info(fmt.Sprintf(
							"received success message for saga submitter '%s': %s", s.id, msg.Json()), common.F(ctx)...)
					case SagaPayloadType_SagaSubmissionFailure:
						common.L.Warn(fmt.Sprintf(
							"received failure message for saga submitter '%s': %s", s.id, msg.Json()), common.F(ctx)...)
					default:
						common.L.Warn(fmt.Sprintf(
							"received unknown message for saga submitter '%s': %s", s.id, msg.Json()), common.F(ctx)...)
					}
					return nil
				},
				func(ctx context.Context, err error) error {
					if err != nil {
						common.L.Warn(fmt.Sprintf(
							"failed to consume messages for saga submitter '%s' on node %s: %v", s.id, inboxNodeName, err), common.F(ctx)...)
					}
					return nil
				},
			)
		}
	}
	s.mu.Unlock()
}

func (s *defaultSagaSubmitter) StartAnnouncement(ctx context.Context) {
	traxCoordinatorBaseUrl := common.GetServiceBaseURL("traxcoord")
	announcementIntervalStr := os.Getenv("TRAX_SUBMITTER_ANNOUNCEMENT_INTERVAL")
	if len(announcementIntervalStr) == 0 {
		panic("TRAX_SUBMITTER_ANNOUNCEMENT_INTERVAL is not set")
	}
	announcementInterval, err := time.ParseDuration(announcementIntervalStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse duration '%s': %v", announcementIntervalStr, err))
	}
	for {
		resp, err := s.announceToCoordinator(traxCoordinatorBaseUrl)
		if err == nil && resp.StatusCode == 200 {
			common.L.Debug(fmt.Sprintf(
				"successfully announced saga submitter '%s'", s.id), common.F(ctx)...)
			s.processAnnouncementResponse(ctx, resp)
			resp.Body.Close()
		} else {
			// Set readyToAcceptSagaSubmissionRequests to false on announcement failure
			s.mu.Lock()
			s.readyToAcceptSagaSubmissionRequests = false
			common.L.Info(fmt.Sprintf("[FAIL] set saga-submitter '%s' as NOT READY", s.id), common.F(ctx)...)
			s.mu.Unlock()

			if err != nil {
				common.L.Warn(fmt.Sprintf(
					"failed to announce saga submitter '%s': %v", s.id, err), common.F(ctx)...)
			} else {
				common.L.Warn(fmt.Sprintf(
					"failed to announce saga submitter '%s': http status > %v [%s]",
					s.id, resp.StatusCode, resp.Status), common.F(ctx)...)
				resp.Body.Close()
			}

			// Fast retry with exponential backoff before falling back to normal interval.
			// This reduces recovery time from ~30s to ~1s for transient coordinator blips.
			backoff := announcementBackoffInitial
			for retry := 0; retry < announcementBackoffMaxRetries; retry++ {
				common.L.Info(fmt.Sprintf(
					"fast-retrying announcement for saga submitter '%s' (attempt %d/%d, backoff %v)",
					s.id, retry+1, announcementBackoffMaxRetries, backoff), common.F(ctx)...)

				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}

				resp, err := s.announceToCoordinator(traxCoordinatorBaseUrl)
				if err == nil && resp.StatusCode == 200 {
					common.L.Info(fmt.Sprintf(
						"saga submitter '%s' recovered on fast-retry attempt %d/%d",
						s.id, retry+1, announcementBackoffMaxRetries), common.F(ctx)...)
					s.processAnnouncementResponse(ctx, resp)
					resp.Body.Close()
					break
				}
				if resp != nil {
					resp.Body.Close()
				}

				// Increase backoff: 1s → 2s → 4s → 8s → 16s (capped at announcementInterval)
				backoff = time.Duration(float64(backoff) * announcementBackoffMultiplier)
				if backoff > announcementInterval {
					backoff = announcementInterval
				}
			}
		}
		time.Sleep(announcementInterval)
	}
}

func (s *defaultSagaSubmitter) SubmitSaga(
	ctx context.Context,
	participantId string,
	traceId string,
	zoneId string,
	origin string,
	originIdempotencyKey string,
	issuer string,
	referrer string,
	tags []string,
	metadata map[string]string,
	sagaTemplateId string,
	sagaInput map[string]string,
) (string, error) {

	clusterId := participantId // TODO(kam): map participantId to clusterId
	if _, ok := s.nodeNames[clusterId]; !ok {
		return "", fmt.Errorf("cluster id not found: %s", clusterId)
	}

	sagaInstanceId := common.SecureRandomString(32)

	outboxNodeName := s.nodeNames[clusterId].Outbox

	messageBuilder := NewTraxMessageBuilder().
		ClusterId(participantId).
		TraceId(traceId).
		Origin(origin).
		Issuer(issuer).
		Referrer(referrer).
		Submitter(s.id).
		Tags(tags).
		Metadata(metadata).
		AnonymousSession() // TODO(kam)

	// Add origin idempotency key if provided
	if originIdempotencyKey != "" {
		messageBuilder = messageBuilder.OriginIdempotencyKey(originIdempotencyKey)
	}

	err := s.mqClient.PublishToNode(
		ctx,
		outboxNodeName,
		string(execpl.ExecutionPipelineMessageTypeEnum_Trax),
		"application/json",
		messageBuilder.
			AddPayload(
				NewPayloadBuilder().
					Type(SagaPayloadType_SagaSubmissionRequest).
					Json(NewSagaSubmissionRequestPayloadBuilder().
						SagaSubmitterId(s.id).
						SagaTemplateId(sagaTemplateId).
						SagaInstanceId(sagaInstanceId).
						ZoneId(zoneId).
						SagaInput(sagaInput).
						Build().
						Json()).
					Build()).
			Build().
			Json(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish saga submission request to node %s: %w", outboxNodeName, err)
	}
	return sagaInstanceId, nil
}

func (s *defaultSagaSubmitter) SubmitSubSaga(
	ctx context.Context,
	participantId string,
	traceId string,
	zoneId string,
	origin string,
	originIdempotencyKey string,
	issuer string,
	referrer string,
	tags []string,
	metadata map[string]string,
	sagaTemplateId string,
	sagaInput map[string]string,
	parentSagaInstanceId string,
	parentSagaStepInstanceId string,
	rootSagaInstanceId string,
	sagaDepth int,
) (string, error) {

	clusterId := participantId // TODO(kam): map participantId to clusterId
	if _, ok := s.nodeNames[clusterId]; !ok {
		return "", fmt.Errorf("cluster id not found: %s", clusterId)
	}

	sagaInstanceId := common.SecureRandomString(32)

	outboxNodeName := s.nodeNames[clusterId].Outbox

	messageBuilder := NewTraxMessageBuilder().
		ClusterId(participantId).
		TraceId(traceId).
		Origin(origin).
		Issuer(issuer).
		Referrer(referrer).
		Submitter(s.id).
		Tags(tags).
		Metadata(metadata).
		AnonymousSession() // TODO(kam)

	if originIdempotencyKey != "" {
		messageBuilder = messageBuilder.OriginIdempotencyKey(originIdempotencyKey)
	}

	err := s.mqClient.PublishToNode(
		ctx,
		outboxNodeName,
		string(execpl.ExecutionPipelineMessageTypeEnum_Trax),
		"application/json",
		messageBuilder.
			AddPayload(
				NewPayloadBuilder().
					Type(SagaPayloadType_SagaSubmissionRequest).
					Json(NewSagaSubmissionRequestPayloadBuilder().
						SagaSubmitterId(s.id).
						SagaTemplateId(sagaTemplateId).
						SagaInstanceId(sagaInstanceId).
						ZoneId(zoneId).
						SagaInput(sagaInput).
						ParentSagaInstanceId(parentSagaInstanceId).
						ParentSagaStepInstanceId(parentSagaStepInstanceId).
						RootSagaInstanceId(rootSagaInstanceId).
						SagaDepth(sagaDepth).
						Build().
						Json()).
					Build()).
			Build().
			Json(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish sub-saga submission request to node %s: %w", outboxNodeName, err)
	}
	return sagaInstanceId, nil
}

// WaitForSagaCompletion polls the traxctrl API until the saga reaches a terminal state
func (s *defaultSagaSubmitter) WaitForSagaCompletion(
	ctx context.Context,
	clusterId string,
	sagaInstanceId string,
	pollInterval time.Duration,
	maxWaitTime time.Duration,
) (*SagaCompletionResult, error) {
	if s.traxCtrlURL == "" {
		return nil, fmt.Errorf("traxCtrlURL not set - call SetTraxCtrlURL first")
	}

	deadline := time.Now().Add(maxWaitTime)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for saga %s to complete after %v", sagaInstanceId, maxWaitTime)
			}

			// Get saga instance status
			sagaInstance, err := s.getSagaInstance(ctx, clusterId, sagaInstanceId)
			if err != nil {
				common.L.Warn(fmt.Sprintf("failed to get saga instance %s: %v, will retry", sagaInstanceId, err))
				continue
			}

			// Check if saga is in a terminal state
			switch sagaInstance.State {
			case SagaStateEnum_Committed:
				// Success - extract outputs from steps
				outputs, err := s.extractOutputsFromSteps(ctx, clusterId, sagaInstanceId)
				if err != nil {
					return nil, fmt.Errorf("saga committed but failed to extract outputs: %w", err)
				}
				return &SagaCompletionResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Committed,
					Outputs:        outputs,
				}, nil

			case SagaStateEnum_Compensated:
				return &SagaCompletionResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Compensated,
					Error:          "saga was compensated (rolled back)",
				}, fmt.Errorf("saga %s was compensated", sagaInstanceId)

			case SagaStateEnum_Blocked:
				return &SagaCompletionResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Blocked,
					Error:          "saga is blocked and requires intervention",
				}, fmt.Errorf("saga %s is blocked", sagaInstanceId)

			case SagaStateEnum_InvalidState:
				return &SagaCompletionResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_InvalidState,
					Error:          "saga entered invalid state",
				}, fmt.Errorf("saga %s entered invalid state", sagaInstanceId)

			case SagaStateEnum_Running:
				// Still running, continue polling
				common.L.Debug(fmt.Sprintf("saga %s still running, continuing to poll", sagaInstanceId))
				continue

			default:
				common.L.Warn(fmt.Sprintf("saga %s in unexpected state: %s", sagaInstanceId, sagaInstance.State))
				continue
			}
		}
	}
}

// getSagaInstance fetches the saga instance from traxctrl API
func (s *defaultSagaSubmitter) getSagaInstance(ctx context.Context, clusterId, sagaInstanceId string) (*SagaInstance, error) {
	url := fmt.Sprintf("%s/saga-instances/%s", s.traxCtrlURL, sagaInstanceId)

	reqBody := struct {
		ClusterId string `json:"cluster_id"`
	}{
		ClusterId: clusterId,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// The API returns a response with JSON-encoded fields
	var apiResp struct {
		InstanceId         string            `json:"instance_id"`
		ClusterId          string            `json:"cluster_id"`
		ZoneId             string            `json:"zone_id"`
		TraceId            string            `json:"trace_id"`
		ExecutionId        string            `json:"execution_id"`
		SagaSubmitterId    string            `json:"saga_submitter_id"`
		Labels             map[string]string `json:"labels"`
		Tags               []string          `json:"tags"`
		Metadata           string            `json:"metadata"`
		State              string            `json:"state"`
		SagaTemplateId     string            `json:"saga_template_id"`
		Input              string            `json:"input_data"`
		SagaInstanceIds    []string          `json:"saga_instance_ids"`
		SagaIdempotencyKey string            `json:"saga_idempotency_key"`
		CreatedAt          int64             `json:"created_at"`
		UpdatedAt          int64             `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to SagaInstance
	sagaInstance := &SagaInstance{
		InstanceId:      apiResp.InstanceId,
		ClusterId:       apiResp.ClusterId,
		ZoneId:          apiResp.ZoneId,
		TraceId:         apiResp.TraceId,
		ExecutionId:     apiResp.ExecutionId,
		SagaSubmitterId: apiResp.SagaSubmitterId,
		Labels:          apiResp.Labels,
		Tags:            apiResp.Tags,
		State:           SagaStateEnum(apiResp.State),
		SagaTemplateId:  apiResp.SagaTemplateId,
		SagaInstanceIds: apiResp.SagaInstanceIds,
		CreatedAt:       apiResp.CreatedAt,
		UpdatedAt:       apiResp.UpdatedAt,
	}

	// Parse metadata if present
	if apiResp.Metadata != "" {
		if err := json.Unmarshal([]byte(apiResp.Metadata), &sagaInstance.Metadata); err != nil {
			common.L.Warn(fmt.Sprintf("failed to parse saga instance metadata: %v", err))
		}
	}

	// Parse input if present
	if apiResp.Input != "" {
		if err := json.Unmarshal([]byte(apiResp.Input), &sagaInstance.Input); err != nil {
			common.L.Warn(fmt.Sprintf("failed to parse saga instance input: %v", err))
		}
	}

	return sagaInstance, nil
}

// extractOutputsFromSteps fetches all step instances and combines their outputs
func (s *defaultSagaSubmitter) extractOutputsFromSteps(ctx context.Context, clusterId, sagaInstanceId string) (map[string]string, error) {
	url := fmt.Sprintf("%s/saga-step-instances/list", s.traxCtrlURL)

	reqBody := struct {
		ClusterId      string `json:"cluster_id"`
		SagaInstanceId string `json:"saga_instance_id"`
	}{
		ClusterId:      clusterId,
		SagaInstanceId: sagaInstanceId,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the list response
	var listResp struct {
		SagaStepInstances []struct {
			InstanceId         string `json:"instance_id"`
			State              string `json:"state"`
			Result             string `json:"result_data"`
			CompensationResult string `json:"compensation_result_data"`
			SagaStepTemplateId string `json:"saga_step_template_id"`
		} `json:"saga_step_instances"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Combine outputs from all successful steps
	outputs := make(map[string]string)
	for _, step := range listResp.SagaStepInstances {
		// Only include outputs from execution-succeeded or execution-done steps
		if step.State != string(SagaStepStateEnum_ExecutionSucceeded) &&
			step.State != string(SagaStepStateEnum_ExecutionDone) {
			continue
		}

		if step.Result == "" {
			continue
		}

		// Parse the result JSON
		var stepOutputs map[string]string
		if err := json.Unmarshal([]byte(step.Result), &stepOutputs); err != nil {
			common.L.Warn(fmt.Sprintf("failed to parse step %s result: %v", step.SagaStepTemplateId, err))
			continue
		}

		// Merge into combined outputs (later steps can override earlier ones)
		for k, v := range stepOutputs {
			outputs[k] = v
		}
	}

	return outputs, nil
}

// ResetForTesting clears the cached cluster state and forces re-announcement.
// This is used during E2E testing when the database is switched dynamically.
func (s *defaultSagaSubmitter) ResetForTesting() {
	s.mu.Lock()
	defer s.mu.Unlock()

	common.L.Info(fmt.Sprintf("Resetting saga submitter '%s' for testing - clearing cluster cache", s.id))

	// Clear the ready state so the next announcement will fully reinitialize
	s.readyToAcceptSagaSubmissionRequests = false

	// Clear cached cluster IDs and node names
	s.clusterIds = []string{}
	s.nodeNames = make(map[string]*SubmitterNodeNames)

	// Note: We do NOT clear consumersStarted because RabbitMQ consumers are still valid
	// and subscribed to the same queue names. The queue names are based on submitter ID,
	// not database state, so existing consumers will continue to work.

	common.L.Info(fmt.Sprintf("Saga submitter '%s' reset complete - will re-announce on next interval", s.id))
}
