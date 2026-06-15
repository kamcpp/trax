package trax

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockSagaSubmitter is a minimal mock for testing SagaContext
type mockSagaSubmitter struct {
	mu                  sync.Mutex
	submitSubSagaCalls  []submitSubSagaCall
	submitSubSagaResult string
	submitSubSagaErr    error
}

type submitSubSagaCall struct {
	ParentSagaInstanceId     string
	ParentSagaStepInstanceId string
	RootSagaInstanceId       string
	SagaDepth                int
	SagaTemplateId           string
	SagaInput                map[string]string
}

func (m *mockSagaSubmitter) Id() string                                  { return "mock" }
func (m *mockSagaSubmitter) StartAnnouncement(ctx context.Context)       {}
func (m *mockSagaSubmitter) IsReadyToAcceptSagaSubmissionRequests() bool { return true }
func (m *mockSagaSubmitter) IsReadyWithClusters() bool                   { return true }
func (m *mockSagaSubmitter) WaitUntilReadyToAcceptSagaSubmissionRequests(ctx context.Context) error {
	return nil
}
func (m *mockSagaSubmitter) GetDefaultClusterId() string { return "test-cluster" }
func (m *mockSagaSubmitter) GetClusterIds() []string     { return []string{"test-cluster"} }
func (m *mockSagaSubmitter) GetTraxCtrlURL() string      { return "http://localhost:17202" }
func (m *mockSagaSubmitter) SetTraxCtrlURL(url string)   {}
func (m *mockSagaSubmitter) ResetForTesting()            {}

func (m *mockSagaSubmitter) SubmitSaga(
	ctx context.Context,
	participantId, traceId, zoneId, origin, originIdempotencyKey, issuer, referrer string,
	tags []string, metadata map[string]string,
	sagaTemplateId string, sagaInput map[string]string,
) (string, error) {
	return "saga-instance-123", nil
}

func (m *mockSagaSubmitter) SubmitSubSaga(
	ctx context.Context,
	participantId, traceId, zoneId, origin, originIdempotencyKey, issuer, referrer string,
	tags []string, metadata map[string]string,
	sagaTemplateId string, sagaInput map[string]string,
	parentSagaInstanceId, parentSagaStepInstanceId, rootSagaInstanceId string,
	sagaDepth int,
) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.submitSubSagaCalls = append(m.submitSubSagaCalls, submitSubSagaCall{
		ParentSagaInstanceId:     parentSagaInstanceId,
		ParentSagaStepInstanceId: parentSagaStepInstanceId,
		RootSagaInstanceId:       rootSagaInstanceId,
		SagaDepth:                sagaDepth,
		SagaTemplateId:           sagaTemplateId,
		SagaInput:                sagaInput,
	})
	return m.submitSubSagaResult, m.submitSubSagaErr
}

func (m *mockSagaSubmitter) WaitForSagaCompletion(
	ctx context.Context, clusterId, sagaInstanceId string,
	pollInterval, maxWaitTime time.Duration,
) (*SagaCompletionResult, error) {
	return nil, nil
}

func TestSagaContextViaContextKey(t *testing.T) {
	// Verify that SagaContext can be stored and retrieved via context.WithValue
	ctx := context.Background()

	// No SagaContext set yet
	sc := GetSagaContext(ctx)
	if sc != nil {
		t.Fatal("expected nil SagaContext from empty context")
	}

	// Set a SagaContext
	sagaCtx := &defaultSagaContext{
		parentSagaInstanceId:     "parent-123",
		parentSagaStepInstanceId: "step-456",
		rootSagaInstanceId:       "root-789",
		sagaDepth:                2,
		clusterId:                "cluster-abc",
	}

	ctx = WithSagaContext(ctx, sagaCtx)
	sc = GetSagaContext(ctx)
	if sc == nil {
		t.Fatal("expected non-nil SagaContext after WithSagaContext")
	}

	if sc.ParentSagaInstanceId() != "parent-123" {
		t.Errorf("ParentSagaInstanceId = %q, want %q", sc.ParentSagaInstanceId(), "parent-123")
	}
	if sc.ParentSagaStepInstanceId() != "step-456" {
		t.Errorf("ParentSagaStepInstanceId = %q, want %q", sc.ParentSagaStepInstanceId(), "step-456")
	}
	if sc.RootSagaInstanceId() != "root-789" {
		t.Errorf("RootSagaInstanceId = %q, want %q", sc.RootSagaInstanceId(), "root-789")
	}
	if sc.SagaDepth() != 2 {
		t.Errorf("SagaDepth = %d, want %d", sc.SagaDepth(), 2)
	}
	if sc.ClusterId() != "cluster-abc" {
		t.Errorf("ClusterId = %q, want %q", sc.ClusterId(), "cluster-abc")
	}
}

func TestSagaContextDoesNotOverrideUnrelatedKeys(t *testing.T) {
	type otherKey struct{}
	ctx := context.WithValue(context.Background(), otherKey{}, "hello")
	ctx = WithSagaContext(ctx, &defaultSagaContext{clusterId: "c1"})

	// SagaContext should be present
	if GetSagaContext(ctx) == nil {
		t.Fatal("SagaContext missing after WithSagaContext")
	}
	// Other key should still be present
	if ctx.Value(otherKey{}) != "hello" {
		t.Fatal("other context key was overridden")
	}
}

func TestDefaultSagaContextAccessors(t *testing.T) {
	sc := &defaultSagaContext{
		parentSagaInstanceId:     "p1",
		parentSagaStepInstanceId: "s1",
		rootSagaInstanceId:       "r1",
		sagaDepth:                0,
		clusterId:                "c1",
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"ParentSagaInstanceId", sc.ParentSagaInstanceId(), "p1"},
		{"ParentSagaStepInstanceId", sc.ParentSagaStepInstanceId(), "s1"},
		{"RootSagaInstanceId", sc.RootSagaInstanceId(), "r1"},
		{"ClusterId", sc.ClusterId(), "c1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
	if sc.SagaDepth() != 0 {
		t.Errorf("SagaDepth = %d, want 0", sc.SagaDepth())
	}
}

func TestSubSagaOptionTimeout(t *testing.T) {
	opts := &subSagaOptions{
		timeout:      10 * time.Minute,
		pollInterval: 2 * time.Second,
	}

	WithSubSagaTimeout(5 * time.Minute)(opts)
	if opts.timeout != 5*time.Minute {
		t.Errorf("timeout = %v, want %v", opts.timeout, 5*time.Minute)
	}

	WithSubSagaPollInterval(500 * time.Millisecond)(opts)
	if opts.pollInterval != 500*time.Millisecond {
		t.Errorf("pollInterval = %v, want %v", opts.pollInterval, 500*time.Millisecond)
	}
}
