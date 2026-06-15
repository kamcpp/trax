package trax

import (
	"encoding/json"
)

type Cluster struct {
	Id          string            `json:"id"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type SagaTemplate struct {
	TemplateId string `json:"template_id"`

	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`

	SagaStepTemplateIds []string `json:"saga_step_template_ids"`
}

type SagaStepTemplate struct {
	TemplateId string `json:"template_id"`

	SagaTemplateId string `json:"saga_template_id"`

	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type SagaInstance struct {
	InstanceId string `json:"instance_id"`

	ClusterId string `json:"cluster_id"`

	ZoneId string `json:"zone_id"`

	// trace id is used by the saga and all its steps
	TraceId string `json:"trace_id"`

	// execution id is different per step and the saga itself
	// has its own execution id
	ExecutionId string `json:"execution_id"`

	SagaSubmitterId string `json:"saga_submitter_id"`

	Origin               string `json:"origin"`
	OriginIdempotencyKey string `json:"origin_idempotency_key"`

	Labels   map[string]string `json:"labels"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`

	State SagaStateEnum `json:"state"`

	SagaTemplateId string `json:"saga_template_id"`

	Input map[string]string `json:"input_data"`

	SagaInstanceIds []string `json:"saga_instance_ids"`

	// Sub-saga hierarchy fields
	ParentSagaInstanceId     string `json:"parent_saga_instance_id"`
	ParentSagaStepInstanceId string `json:"parent_saga_step_instance_id"`
	RootSagaInstanceId       string `json:"root_saga_instance_id"`
	SagaDepth                int    `json:"saga_depth"`

	CompensationReason string `json:"compensation_reason"`

	// AnnexIids identifies binary attachments stored alongside this
	// saga in the trax `saga_annexes` table. Populated by the saga
	// submitter (e.g. csdmsggw uploading the bytes of a batch-issue-
	// security-units payload after the saga is created); consumed
	// by sd_admin's saga watcher and by partner-facing endpoints
	// that need to surface the originating annexes.
	AnnexIids []string `json:"annex_iids"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// SagaAnnex is a byte-content attachment tied to a SagaInstance
// (by `saga_instance_id`). Storage lives in the trax `saga_annexes`
// table — trax is the owner; csdmsggw and other gateways write
// into trax via traxctrl's HTTP endpoints rather than maintaining
// their own annex tables.
type SagaAnnex struct {
	Iid            string `json:"iid"`
	ClusterId      string `json:"cluster_id"`
	SagaInstanceId string `json:"saga_instance_id"`
	ContentType    string `json:"content_type"`
	ContentLength  int64  `json:"content_length"`
	Notes          string `json:"notes,omitempty"`
	// ContentData carries the raw bytes — only the GetSagaAnnexBytes
	// path returns it; List/Create do not. Keeping it on the same
	// struct avoids a second type for every layer.
	ContentData []byte `json:"-"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type SagaStepExecutionLog struct {
	NextExecutionTs           int64 `json:"next_execution_ts"`
	ExecutionRequestSentTs    int64 `json:"execution_request_sent_ts"`
	ExecutionTimeoutTs        int64 `json:"execution_timeout_ts"`
	ExecutionResultReceivedTs int64 `json:"execution_result_received_ts"`
	LogConclusionTs           int64 `json:"log_conclusion_ts"`

	ExecutionResult map[string]string `json:"execution_result"`
	ExecutionError  string            `json:"execution_error"`

	IsCompensation bool              `json:"is_compensation"`
	Metadata       map[string]string `json:"metadata"`
}

type SagaStepInstance struct {
	InstanceId string `json:"instance_id"`

	ClusterId string `json:"cluster_id"`

	ZoneId string `json:"zone_id"`

	SagaInstanceId string `json:"saga_instance_id"`

	// trace id is used by the saga and all its steps
	TraceId string `json:"trace_id"`

	// execution id is different per step and the saga itself
	// has its own execution id
	ExecutionId string `json:"execution_id"`

	Labels   map[string]string `json:"labels"`
	Tags     []string          `json:"tags"`
	Metadata map[string]string `json:"metadata"`

	Affinity string `json:"affinity"`

	State SagaStepStateEnum `json:"state"`

	Result             map[string]string `json:"result_data"`
	CompensationResult map[string]string `json:"compensation_result_data"`

	SagaTemplateId     string `json:"saga_template_id"`
	SagaStepTemplateId string `json:"saga_step_template_id"`

	PreviousSagaStepInstanceId string `json:"previous_saga_step_instance_id"`
	NextSagaStepInstanceId     string `json:"next_saga_step_instance_id"`

	ExecutionHistory []*SagaStepExecutionLog `json:"execution_history"`

	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

func (si *SagaInstance) SagaIdempotencyKey() string {
	return getSagaIdempotencyKey(
		si.ClusterId,
		si.ZoneId,
		si.SagaTemplateId,
		si.InstanceId,
	)
}

func (si *SagaInstance) Json() string {
	data, err := json.Marshal(si)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (ssi *SagaStepInstance) SagaIdempotencyKey() string {
	return getSagaStepIdempotencyKey(
		ssi.ClusterId,
		ssi.ZoneId,
		ssi.SagaTemplateId,
		ssi.SagaStepTemplateId,
		ssi.InstanceId,
	)
}

func (ssi *SagaStepInstance) Json() string {
	data, err := json.Marshal(ssi)
	if err != nil {
		panic(err)
	}
	return string(data)
}
