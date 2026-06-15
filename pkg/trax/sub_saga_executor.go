package trax

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kamcpp/trax/pkg/common"
)

// SubSagaResult contains the result of a sub-saga execution
type SubSagaResult struct {
	// SagaInstanceId is the ID of the spawned sub-saga
	SagaInstanceId string

	// State is the final state of the sub-saga
	State SagaStateEnum

	// Outputs contains the combined outputs from all successful steps
	Outputs map[string]string

	// Error contains any error message if the saga failed
	Error string
}

// SubSagaExecutor handles spawning and waiting for sub-sagas to complete
type SubSagaExecutor struct {
	sagaSubmitter SagaSubmitter
	traxCtrlURL   string
	pollInterval  time.Duration
	maxWaitTime   time.Duration
	httpClient    *http.Client
}

// SubSagaExecutorOption is a functional option for configuring SubSagaExecutor
type SubSagaExecutorOption func(*SubSagaExecutor)

// WithPollInterval sets the polling interval for checking sub-saga status
func WithPollInterval(interval time.Duration) SubSagaExecutorOption {
	return func(e *SubSagaExecutor) {
		e.pollInterval = interval
	}
}

// WithMaxWaitTime sets the maximum time to wait for sub-saga completion
func WithMaxWaitTime(maxWait time.Duration) SubSagaExecutorOption {
	return func(e *SubSagaExecutor) {
		e.maxWaitTime = maxWait
	}
}

// WithHTTPClient sets a custom HTTP client for API calls
func WithHTTPClient(client *http.Client) SubSagaExecutorOption {
	return func(e *SubSagaExecutor) {
		e.httpClient = client
	}
}

// NewSubSagaExecutor creates a new SubSagaExecutor
func NewSubSagaExecutor(sagaSubmitter SagaSubmitter, traxCtrlURL string, opts ...SubSagaExecutorOption) *SubSagaExecutor {
	e := &SubSagaExecutor{
		sagaSubmitter: sagaSubmitter,
		traxCtrlURL:   traxCtrlURL,
		pollInterval:  2 * time.Second,
		maxWaitTime:   10 * time.Minute,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// SpawnAndWaitWithParent submits a sub-saga with parent context and waits for completion.
// The parent context fields are included in the submission payload so the coordinator
// registers the parent-child relationship in the database.
func (e *SubSagaExecutor) SpawnAndWaitWithParent(
	ctx context.Context,
	clusterId string,
	sagaTemplateId string,
	sagaInput map[string]string,
	originIdempotencyKey string,
	parentSagaInstanceId string,
	parentSagaStepInstanceId string,
	rootSagaInstanceId string,
	parentSagaDepth int,
) (*SubSagaResult, error) {
	common.L.Info(fmt.Sprintf(
		"[SUB-SAGA] submitting sub-saga: template=%s, cluster=%s, originKey=%s, parent=%s, parentStep=%s, root=%s, depth=%d",
		sagaTemplateId, clusterId, originIdempotencyKey,
		parentSagaInstanceId, parentSagaStepInstanceId, rootSagaInstanceId, parentSagaDepth+1))

	sagaInstanceId, err := e.sagaSubmitter.SubmitSubSaga(
		ctx,
		clusterId,
		"",             // traceId
		"default-zone", // zoneId
		"sub-saga",     // origin
		originIdempotencyKey,
		"",  // issuer
		"",  // referrer
		nil, // tags
		nil, // metadata
		sagaTemplateId,
		sagaInput,
		parentSagaInstanceId,
		parentSagaStepInstanceId,
		rootSagaInstanceId,
		parentSagaDepth+1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to submit sub-saga %s (cluster=%s): %w", sagaTemplateId, clusterId, err)
	}

	common.L.Info(fmt.Sprintf(
		"[SUB-SAGA] submitted successfully: template=%s, instanceId=%s, cluster=%s (parent=%s, root=%s, depth=%d)",
		sagaTemplateId, sagaInstanceId, clusterId, parentSagaInstanceId, rootSagaInstanceId, parentSagaDepth+1))

	common.L.Info(fmt.Sprintf(
		"[SUB-SAGA] starting poll for completion: instanceId=%s, cluster=%s, traxCtrlURL=%s, pollInterval=%v, maxWait=%v",
		sagaInstanceId, clusterId, e.traxCtrlURL, e.pollInterval, e.maxWaitTime))

	result, err := e.pollForCompletion(ctx, clusterId, sagaInstanceId)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for sub-saga %s completion (cluster=%s, traxCtrlURL=%s): %w",
			sagaInstanceId, clusterId, e.traxCtrlURL, err)
	}

	return result, nil
}

// pollForCompletion polls the traxctrl API until the saga reaches a terminal state.
// It implements early failure detection: if the saga instance is not found after
// maxConsecutiveErrors consecutive poll attempts, it fails immediately rather than
// waiting for the full timeout. This surfaces issues like failed saga creation quickly.
func (e *SubSagaExecutor) pollForCompletion(ctx context.Context, clusterId, sagaInstanceId string) (*SubSagaResult, error) {
	deadline := time.Now().Add(e.maxWaitTime)
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	const maxConsecutiveErrors = 15 // ~30 seconds with 2s poll interval
	consecutiveErrors := 0
	var lastError error
	everFoundSaga := false

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for sub-saga %s to complete after %v (lastError: %v)", sagaInstanceId, e.maxWaitTime, lastError)
			}

			// Get saga instance status
			sagaInstance, err := e.getSagaInstance(ctx, clusterId, sagaInstanceId)
			if err != nil {
				consecutiveErrors++
				lastError = err
				common.L.Warn(fmt.Sprintf(
					"failed to get saga instance %s (attempt %d/%d): %v",
					sagaInstanceId, consecutiveErrors, maxConsecutiveErrors, err))

				// Early failure: if we've never seen the saga and hit too many consecutive errors,
				// the saga was likely never created (e.g., coordinator rejected the submission).
				if !everFoundSaga && consecutiveErrors >= maxConsecutiveErrors {
					return nil, fmt.Errorf(
						"sub-saga %s not found after %d consecutive poll attempts (cluster=%s). "+
							"The saga was likely never created by the coordinator. Last error: %w",
						sagaInstanceId, consecutiveErrors, clusterId, lastError)
				}
				continue
			}

			// Reset error counter on successful poll
			consecutiveErrors = 0
			everFoundSaga = true

			// Check if saga is in a terminal state
			switch sagaInstance.State {
			case SagaStateEnum_Committed:
				// Success - extract outputs from steps
				outputs, err := e.extractOutputsFromSteps(ctx, clusterId, sagaInstanceId)
				if err != nil {
					return nil, fmt.Errorf("saga committed but failed to extract outputs: %w", err)
				}
				return &SubSagaResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Committed,
					Outputs:        outputs,
				}, nil

			case SagaStateEnum_Compensated:
				compOutputs, compErr := e.extractCompensationOutputsFromSteps(ctx, clusterId, sagaInstanceId)
				if compErr != nil {
					common.L.Warn(fmt.Sprintf("failed to extract compensation outputs from sub-saga %s: %v", sagaInstanceId, compErr))
				}
				return &SubSagaResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Compensated,
					Outputs:        compOutputs,
					Error:          "sub-saga was compensated (rolled back)",
				}, fmt.Errorf("sub-saga %s was compensated", sagaInstanceId)

			case SagaStateEnum_Blocked:
				return &SubSagaResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_Blocked,
					Error:          "sub-saga is blocked and requires intervention",
				}, fmt.Errorf("sub-saga %s is blocked", sagaInstanceId)

			case SagaStateEnum_InvalidState:
				return &SubSagaResult{
					SagaInstanceId: sagaInstanceId,
					State:          SagaStateEnum_InvalidState,
					Error:          "sub-saga entered invalid state",
				}, fmt.Errorf("sub-saga %s entered invalid state", sagaInstanceId)

			case SagaStateEnum_Running:
				// Still running, continue polling
				common.L.Debug(fmt.Sprintf("sub-saga %s still running, continuing to poll", sagaInstanceId))
				continue

			default:
				common.L.Warn(fmt.Sprintf("sub-saga %s in unexpected state: %s", sagaInstanceId, sagaInstance.State))
				continue
			}
		}
	}
}

// getSagaInstance fetches the saga instance from traxctrl API
func (e *SubSagaExecutor) getSagaInstance(ctx context.Context, clusterId, sagaInstanceId string) (*SagaInstance, error) {
	url := fmt.Sprintf("%s/saga-instances/%s", e.traxCtrlURL, sagaInstanceId)

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

	resp, err := e.httpClient.Do(req)
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
			// Ignore metadata parsing errors
			common.L.Warn(fmt.Sprintf("failed to parse saga instance metadata: %v", err))
		}
	}

	// Parse input if present
	if apiResp.Input != "" {
		if err := json.Unmarshal([]byte(apiResp.Input), &sagaInstance.Input); err != nil {
			// Ignore input parsing errors
			common.L.Warn(fmt.Sprintf("failed to parse saga instance input: %v", err))
		}
	}

	return sagaInstance, nil
}

// extractOutputsFromSteps fetches all step instances and combines their outputs
func (e *SubSagaExecutor) extractOutputsFromSteps(ctx context.Context, clusterId, sagaInstanceId string) (map[string]string, error) {
	url := fmt.Sprintf("%s/saga-step-instances/list", e.traxCtrlURL)

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

	resp, err := e.httpClient.Do(req)
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

// extractCompensationOutputsFromSteps fetches all step instances and combines their compensation outputs
func (e *SubSagaExecutor) extractCompensationOutputsFromSteps(ctx context.Context, clusterId, sagaInstanceId string) (map[string]string, error) {
	url := fmt.Sprintf("%s/saga-step-instances/list", e.traxCtrlURL)

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

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp struct {
		SagaStepInstances []struct {
			InstanceId         string `json:"instance_id"`
			State              string `json:"state"`
			CompensationResult string `json:"compensation_result_data"`
			SagaStepTemplateId string `json:"saga_step_template_id"`
		} `json:"saga_step_instances"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	outputs := make(map[string]string)
	for _, step := range listResp.SagaStepInstances {
		if step.State != string(SagaStepStateEnum_CompensationDone) {
			continue
		}

		if step.CompensationResult == "" {
			continue
		}

		var stepOutputs map[string]string
		if err := json.Unmarshal([]byte(step.CompensationResult), &stepOutputs); err != nil {
			common.L.Warn(fmt.Sprintf("failed to parse step %s compensation result: %v", step.SagaStepTemplateId, err))
			continue
		}

		for k, v := range stepOutputs {
			outputs[k] = v
		}
	}

	return outputs, nil
}
