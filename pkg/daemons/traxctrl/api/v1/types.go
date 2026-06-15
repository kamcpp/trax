package apiv1

type sagaTemplateResponse struct {
	TemplateId          string                     `json:"template_id"`
	DisplayName         string                     `json:"display_name"`
	Description         string                     `json:"description"`
	Labels              map[string]string          `json:"labels"`
	Tags                []string                   `json:"tags"`
	Metadata            string                     `json:"metadata"`
	SagaStepTemplateIds []string                   `json:"saga_step_template_ids"`
	SagaStepTemplates   []sagaStepTemplateResponse `json:"saga_step_templates"`
}

type sagaStepTemplateResponse struct {
	TemplateId     string            `json:"template_id"`
	SagaTemplateId string            `json:"saga_template_id"`
	DisplayName    string            `json:"display_name"`
	Description    string            `json:"description"`
	Labels         map[string]string `json:"labels"`
	Tags           []string          `json:"tags"`
	Metadata       string            `json:"metadata"`
}

type listSagaTemplatesResponse struct {
	SagaTemplates []sagaTemplateResponse `json:"saga_templates"`
}

type listSagaTemplateIdsResponse struct {
	SagaTemplateIds []string `json:"saga_template_ids"`
}

// Saga Instance response types
type sagaInstanceResponse struct {
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
	// Sub-saga hierarchy fields
	ParentSagaInstanceId     string `json:"parent_saga_instance_id,omitempty"`
	ParentSagaStepInstanceId string `json:"parent_saga_step_instance_id,omitempty"`
	RootSagaInstanceId       string `json:"root_saga_instance_id,omitempty"`
	SagaDepth                int    `json:"saga_depth"`
	CompensationReason       string `json:"compensation_reason,omitempty"`
	// Saga annex iids — populated by gateways (csdmsggw, …) when
	// they upload binary attachments after submitting the saga.
	AnnexIids []string `json:"annex_iids,omitempty"`
}

type listSagaInstancesResponse struct {
	SagaInstances []sagaInstanceResponse `json:"saga_instances"`
	TotalCount    *int                   `json:"total_count,omitempty"`
}

type listSagaInstanceIdsResponse struct {
	SagaInstanceIds []string `json:"saga_instance_ids"`
}

// Saga Step Instance response types
type sagaStepInstanceResponse struct {
	InstanceId                 string            `json:"instance_id"`
	ClusterId                  string            `json:"cluster_id"`
	ZoneId                     string            `json:"zone_id"`
	SagaInstanceId             string            `json:"saga_instance_id"`
	TraceId                    string            `json:"trace_id"`
	ExecutionId                string            `json:"execution_id"`
	Labels                     map[string]string `json:"labels"`
	Tags                       []string          `json:"tags"`
	Metadata                   string            `json:"metadata"`
	Affinity                   string            `json:"affinity"`
	State                      string            `json:"state"`
	Result                     string            `json:"result_data"`
	CompensationResult         string            `json:"compensation_result_data"`
	SagaTemplateId             string            `json:"saga_template_id"`
	SagaStepTemplateId         string            `json:"saga_step_template_id"`
	PreviousSagaStepInstanceId string            `json:"previous_saga_step_instance_id"`
	NextSagaStepInstanceId     string            `json:"next_saga_step_instance_id"`
	ExecutionHistory           string            `json:"execution_history"`
	SagaIdempotencyKey         string            `json:"saga_idempotency_key"`
	ExecutionError             string            `json:"execution_error,omitempty"`
}

type listSagaStepInstancesResponse struct {
	SagaStepInstances []sagaStepInstanceResponse `json:"saga_step_instances"`
}

type listSagaStepInstanceIdsResponse struct {
	SagaStepInstanceIds []string `json:"saga_step_instance_ids"`
}

// sortFieldRequest mirrors common.SortField on the wire (REST/JSON).
// Direction is "asc" or "desc" — case-insensitive. Empty direction is ASC.
type sortFieldRequest struct {
	Field     string `json:"field"`
	Direction string `json:"direction,omitempty"`
}

// Request types for saga instances. The listing contract:
//   - PageNr / PageSize for pagination (defaults: 1 / 50, max page_size 500).
//   - Search runs ILIKE across every visible text + JSONB column on saga_instances.
//   - SortFields is a multi-column ORDER BY; empty falls back to the store's
//     default (`created_at DESC, instance_id ASC`).
//   - State / SagaTemplateId / SagaSubmitterId are exact-match filters used
//     by the saga-admin tooling (brkadmsvc, prtagent). Each maps to
//     QueryOptions.Filters with the corresponding SQL column name.
//
// See docs/TODO_UNIFORM_LISTING_CONTRACT.md.
type listSagaInstancesRequest struct {
	ClusterId       string             `json:"cluster_id" binding:"required"`
	PageNr          *int               `json:"page_nr,omitempty"`
	PageSize        *int               `json:"page_size,omitempty"`
	Search          string             `json:"search,omitempty"`
	SortFields      []sortFieldRequest `json:"sort_fields,omitempty"`
	State           string             `json:"state,omitempty"`
	SagaTemplateId  string             `json:"saga_template_id,omitempty"`
	SagaSubmitterId string             `json:"saga_submitter_id,omitempty"`
}

type listSagaInstanceIdsRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

type getSagaInstanceRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

type getSagaInstanceChildrenRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

type getSagaInstanceTreeRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

type sagaInstanceTreeResponse struct {
	SagaInstances []sagaInstanceResponse `json:"saga_instances"`
}

// Request types for saga step instances
type listSagaStepInstancesRequest struct {
	ClusterId      string `json:"cluster_id" binding:"required"`
	SagaInstanceId string `json:"saga_instance_id,omitempty"`
}

type listSagaStepInstanceIdsRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

type getSagaStepInstanceRequest struct {
	ClusterId string `json:"cluster_id" binding:"required"`
}

// Cluster response types
type clusterResponse struct {
	Id          string            `json:"id"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    string            `json:"metadata"`
}

type listClustersResponse struct {
	Clusters []clusterResponse `json:"clusters"`
}

type listClusterIdsResponse struct {
	ClusterIds []string `json:"cluster_ids"`
}

// Request types for clusters
type createClusterRequest struct {
	Id          string            `json:"id" binding:"required"`
	DisplayName string            `json:"display_name" binding:"required"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    string            `json:"metadata,omitempty"`
}

type updateClusterRequest struct {
	DisplayName string            `json:"display_name" binding:"required"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    string            `json:"metadata,omitempty"`
}

// Request types for saga template management
type updateSagaTemplateRequest struct {
	DisplayName         string            `json:"display_name"`
	Description         string            `json:"description,omitempty"`
	Labels              map[string]string `json:"labels,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	SagaStepTemplateIds []string          `json:"saga_step_template_ids"`
}

type updateSagaStepTemplateRequest struct {
	SagaTemplateId string            `json:"saga_template_id" binding:"required"`
	DisplayName    string            `json:"display_name"`
	Description    string            `json:"description,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}
