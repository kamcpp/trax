package trax

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CoordinatorProxy interface {
	GetSagaStepTemplate(
		ctx context.Context,
		affinity string,
		tenantId, zoneId, sagaId, sagaStepId string,
	) (*SagaStepTemplate, error)
}

type restCoordinatorProxy struct {
	baseURL string
}

func NewRestCoordinatorProxy(
	baseURL string,
) CoordinatorProxy {
	return &restCoordinatorProxy{
		baseURL: baseURL,
	}
}

func (c restCoordinatorProxy) GetSagaStepTemplate(
	ctx context.Context,
	affinity string,
	tenantId, zoneId, sagaId, sagaStepId string,
) (*SagaStepTemplate, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest(
		"GET", fmt.Sprintf("%s/sagas/%s/steps/%s",
			c.baseURL, sagaId, sagaStepId), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get saga step template: %s", resp.Status)
	}
	var template SagaStepTemplate
	if err := json.NewDecoder(resp.Body).
		Decode(&template); err != nil {
		return nil, err
	}
	return &template, nil
}
