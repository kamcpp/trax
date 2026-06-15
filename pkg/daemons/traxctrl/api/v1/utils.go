package apiv1

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/trax"
)

func convertSagaTemplateToResponse(c *gin.Context, sagaTemplate *trax.SagaTemplate) (*sagaTemplateResponse, error) {
	var stepTemplates []sagaStepTemplateResponse

	// Get step templates
	for _, stepTemplateId := range sagaTemplate.SagaStepTemplateIds {
		stepTemplate, err := traxStore.GetSagaStepTemplate(c, stepTemplateId)
		if err != nil {
			return nil, fmt.Errorf("failed to get step template %q: %v", stepTemplateId, err)
		}

		stepMetadataJSON, err := json.Marshal(stepTemplate.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal step template metadata: %v", err)
		}

		stepResp := sagaStepTemplateResponse{
			TemplateId:     stepTemplate.TemplateId,
			SagaTemplateId: stepTemplate.SagaTemplateId,
			DisplayName:    stepTemplate.DisplayName,
			Description:    stepTemplate.Description,
			Labels:         stepTemplate.Labels,
			Tags:           stepTemplate.Tags,
			Metadata:       string(stepMetadataJSON),
		}
		stepTemplates = append(stepTemplates, stepResp)
	}

	sagaMetadataJSON, err := json.Marshal(sagaTemplate.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga template metadata: %v", err)
	}

	templateResp := &sagaTemplateResponse{
		TemplateId:          sagaTemplate.TemplateId,
		DisplayName:         sagaTemplate.DisplayName,
		Description:         sagaTemplate.Description,
		Labels:              sagaTemplate.Labels,
		Tags:                sagaTemplate.Tags,
		Metadata:            string(sagaMetadataJSON),
		SagaStepTemplateIds: sagaTemplate.SagaStepTemplateIds,
		SagaStepTemplates:   stepTemplates,
	}

	return templateResp, nil
}

func convertSagaInstanceToResponse(sagaInstance *trax.SagaInstance) (*sagaInstanceResponse, error) {
	metadataJSON, err := json.Marshal(sagaInstance.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga instance metadata: %v", err)
	}

	inputJSON, err := json.Marshal(sagaInstance.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga instance input: %v", err)
	}

	return &sagaInstanceResponse{
		InstanceId:         sagaInstance.InstanceId,
		ClusterId:          sagaInstance.ClusterId,
		ZoneId:             sagaInstance.ZoneId,
		TraceId:            sagaInstance.TraceId,
		ExecutionId:        sagaInstance.ExecutionId,
		SagaSubmitterId:    sagaInstance.SagaSubmitterId,
		Labels:             sagaInstance.Labels,
		Tags:               sagaInstance.Tags,
		Metadata:           string(metadataJSON),
		State:              string(sagaInstance.State),
		SagaTemplateId:     sagaInstance.SagaTemplateId,
		Input:              string(inputJSON),
		SagaInstanceIds:    sagaInstance.SagaInstanceIds,
		SagaIdempotencyKey: sagaInstance.SagaIdempotencyKey(),
		CreatedAt:          sagaInstance.CreatedAt,
		UpdatedAt:          sagaInstance.UpdatedAt,
		// Sub-saga hierarchy
		ParentSagaInstanceId:     sagaInstance.ParentSagaInstanceId,
		ParentSagaStepInstanceId: sagaInstance.ParentSagaStepInstanceId,
		RootSagaInstanceId:       sagaInstance.RootSagaInstanceId,
		SagaDepth:                sagaInstance.SagaDepth,
		CompensationReason:       sagaInstance.CompensationReason,
		AnnexIids:                sagaInstance.AnnexIids,
	}, nil
}

func convertSagaStepInstanceToResponse(sagaStepInstance *trax.SagaStepInstance) (*sagaStepInstanceResponse, error) {
	metadataJSON, err := json.Marshal(sagaStepInstance.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga step instance metadata: %v", err)
	}

	executionHistoryJSON, err := json.Marshal(sagaStepInstance.ExecutionHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga step instance execution history: %v", err)
	}

	resultJSON, err := json.Marshal(sagaStepInstance.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga step instance result: %v", err)
	}

	compensationResultJSON, err := json.Marshal(sagaStepInstance.CompensationResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal saga step instance compensation result: %v", err)
	}

	// Extract execution error from execution history (first non-empty error found)
	// Check both ExecutionError field and ExecutionResult["error"] since the TRAX
	// executor framework stores errors in ExecutionResult["error"]
	executionError := ""
	for _, log := range sagaStepInstance.ExecutionHistory {
		if log.ExecutionError != "" {
			executionError = log.ExecutionError
			break
		}
		// Also check ExecutionResult["error"] - this is where the TRAX executor stores errors
		if errMsg, ok := log.ExecutionResult["error"]; ok && errMsg != "" {
			executionError = errMsg
			break
		}
	}

	return &sagaStepInstanceResponse{
		InstanceId:                 sagaStepInstance.InstanceId,
		ClusterId:                  sagaStepInstance.ClusterId,
		ZoneId:                     sagaStepInstance.ZoneId,
		SagaInstanceId:             sagaStepInstance.SagaInstanceId,
		TraceId:                    sagaStepInstance.TraceId,
		ExecutionId:                sagaStepInstance.ExecutionId,
		Labels:                     sagaStepInstance.Labels,
		Tags:                       sagaStepInstance.Tags,
		Metadata:                   string(metadataJSON),
		Affinity:                   sagaStepInstance.Affinity,
		State:                      string(sagaStepInstance.State),
		Result:                     string(resultJSON),
		CompensationResult:         string(compensationResultJSON),
		SagaTemplateId:             sagaStepInstance.SagaTemplateId,
		SagaStepTemplateId:         sagaStepInstance.SagaStepTemplateId,
		PreviousSagaStepInstanceId: sagaStepInstance.PreviousSagaStepInstanceId,
		NextSagaStepInstanceId:     sagaStepInstance.NextSagaStepInstanceId,
		ExecutionHistory:           string(executionHistoryJSON),
		SagaIdempotencyKey:         sagaStepInstance.SagaIdempotencyKey(),
		ExecutionError:             executionError,
	}, nil
}

func convertClusterToResponse(cluster *trax.Cluster) (*clusterResponse, error) {
	metadataJSON, err := json.Marshal(cluster.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cluster metadata: %v", err)
	}

	return &clusterResponse{
		Id:          cluster.Id,
		DisplayName: cluster.DisplayName,
		Description: cluster.Description,
		Labels:      cluster.Labels,
		Tags:        cluster.Tags,
		Metadata:    string(metadataJSON),
	}, nil
}

func convertCreateRequestToCluster(req *createClusterRequest) (*trax.Cluster, error) {
	labels := req.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var metadata map[string]string
	if req.Metadata != "" {
		err := json.Unmarshal([]byte(req.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
		}
	} else {
		metadata = make(map[string]string)
	}

	return &trax.Cluster{
		Id:          req.Id,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Labels:      labels,
		Tags:        tags,
		Metadata:    metadata,
	}, nil
}

func convertUpdateRequestToCluster(clusterId string, req *updateClusterRequest) (*trax.Cluster, error) {
	labels := req.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var metadata map[string]string
	if req.Metadata != "" {
		err := json.Unmarshal([]byte(req.Metadata), &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
		}
	} else {
		metadata = make(map[string]string)
	}

	return &trax.Cluster{
		Id:          clusterId,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Labels:      labels,
		Tags:        tags,
		Metadata:    metadata,
	}, nil
}
