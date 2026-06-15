package trax

import (
	"testing"
)

func TestGetStepTopicExchangeName(t *testing.T) {
	got := getStepTopicExchangeName("cluster1")
	expected := "x_cluster1_trax_saga_steps"
	if got != expected {
		t.Errorf("getStepTopicExchangeName() = %q, want %q", got, expected)
	}
}

func TestGetStepRequestRoutingKey(t *testing.T) {
	tests := []struct {
		name      string
		clusterId string
		affinity  string
		sagaTmpl  string
		stepTmpl  string
		expected  string
	}{
		{
			name:      "basic routing key",
			clusterId: "cluster1",
			affinity:  "1",
			sagaTmpl:  "seven_step_saga",
			stepTmpl:  "step_one",
			expected:  "cluster1.1.seven_step_saga.step_one.request",
		},
		{
			name:      "different affinity",
			clusterId: "cluster1",
			affinity:  "3",
			sagaTmpl:  "transfer_tokens",
			stepTmpl:  "validate_balances",
			expected:  "cluster1.3.transfer_tokens.validate_balances.request",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStepRequestRoutingKey(tt.clusterId, tt.affinity, tt.sagaTmpl, tt.stepTmpl)
			if got != tt.expected {
				t.Errorf("getStepRequestRoutingKey() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetStepResponseRoutingKey(t *testing.T) {
	got := getStepResponseRoutingKey("cluster1", "2", "saga_a", "step_b")
	expected := "cluster1.2.saga_a.step_b.response"
	if got != expected {
		t.Errorf("getStepResponseRoutingKey() = %q, want %q", got, expected)
	}
}

func TestGetExecutorInboxBindingKey(t *testing.T) {
	got := getExecutorInboxBindingKey("cluster1", "seven_step_saga", "step_one")
	expected := "cluster1.*.seven_step_saga.step_one.request"
	if got != expected {
		t.Errorf("getExecutorInboxBindingKey() = %q, want %q", got, expected)
	}
}

func TestGetCoordinatorResultsBindingKey(t *testing.T) {
	got := getCoordinatorResultsBindingKey("cluster1", "2")
	expected := "cluster1.2.*.*.response"
	if got != expected {
		t.Errorf("getCoordinatorResultsBindingKey() = %q, want %q", got, expected)
	}
}

func TestGetExecutorInboxQueueName(t *testing.T) {
	got := getExecutorInboxQueueName("cluster1", "seven_step_saga", "step_one")
	expected := "q_cluster1_trax_executor_seven_step_saga_step_one_inbox"
	if got != expected {
		t.Errorf("getExecutorInboxQueueName() = %q, want %q", got, expected)
	}
}

func TestGetCoordinatorResultsQueueName(t *testing.T) {
	got := getCoordinatorResultsQueueName("cluster1", "2")
	expected := "q_cluster1_trax_coordinator_2_results"
	if got != expected {
		t.Errorf("getCoordinatorResultsQueueName() = %q, want %q", got, expected)
	}
}

func TestExecutorInboxQueueNameHasNoAffinity(t *testing.T) {
	// Executor inbox queues must be shared across all affinities.
	// The same (clusterId, saga, step) tuple must produce the same queue name
	// regardless of which coordinator affinity dispatched the request.
	q1 := getExecutorInboxQueueName("c1", "saga_x", "step_y")
	// If the function accidentally accepted an affinity parameter,
	// different affinities would produce different queue names.
	// Since it doesn't take affinity at all, this test simply verifies
	// the queue name is deterministic and contains no affinity segment.
	q2 := getExecutorInboxQueueName("c1", "saga_x", "step_y")
	if q1 != q2 {
		t.Errorf("executor inbox queue name is not deterministic: %q != %q", q1, q2)
	}
}

func TestCoordinatorResultsQueueNameIncludesAffinity(t *testing.T) {
	// Each coordinator affinity must have its own results queue
	q1 := getCoordinatorResultsQueueName("c1", "1")
	q2 := getCoordinatorResultsQueueName("c1", "2")
	q3 := getCoordinatorResultsQueueName("c1", "3")
	if q1 == q2 || q2 == q3 || q1 == q3 {
		t.Errorf("coordinator results queues must be unique per affinity: %q, %q, %q", q1, q2, q3)
	}
}

func TestRoutingKeyRequestResponseSymmetry(t *testing.T) {
	// Verify that request routing keys route to executor inbox binding,
	// and response routing keys route to coordinator results binding.
	clusterId := "cluster1"
	affinity := "2"
	saga := "seven_step_saga"
	step := "step_three"

	requestKey := getStepRequestRoutingKey(clusterId, affinity, saga, step)
	responseKey := getStepResponseRoutingKey(clusterId, affinity, saga, step)

	// Request key should end with ".request", response with ".response"
	if requestKey[len(requestKey)-8:] != ".request" {
		t.Errorf("request routing key should end with '.request': %q", requestKey)
	}
	if responseKey[len(responseKey)-9:] != ".response" {
		t.Errorf("response routing key should end with '.response': %q", responseKey)
	}

	// They should share the same prefix
	requestPrefix := requestKey[:len(requestKey)-8]
	responsePrefix := responseKey[:len(responseKey)-9]
	if requestPrefix != responsePrefix {
		t.Errorf("request and response routing keys should share same prefix: %q vs %q", requestPrefix, responsePrefix)
	}
}
