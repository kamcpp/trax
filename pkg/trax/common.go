package trax

import "fmt"

func getSagaIdempotencyKey(clusterId, zoneId, sagaTemplateId, sagaInstanceId string) string {
	return fmt.Sprintf("sidk:%s.%s.%s.%s",
		clusterId, zoneId, sagaTemplateId, sagaInstanceId)
}

func getSagaStepIdempotencyKey(clusterId, zoneId, sagaTemplateId, sagaStepTemplateId, sagaInstanceId string) string {
	return fmt.Sprintf("ssidk:%s.%s.%s.%s.%s",
		clusterId, zoneId, sagaTemplateId, sagaStepTemplateId, sagaInstanceId)
}

func getSagaSubmitterInboxNodeNames(mqClient MQClient, clusterId, sagaSubmitterId string) (string, string) {
	key := fmt.Sprintf("%s_traxcoord_saga_submitter_%s_inbox", clusterId, sagaSubmitterId)
	return mqClient.GetPublishNodeName(key), mqClient.GetSubscribeNodeName(key)
}

func getSagaSubmitterOutboxNodeNames(mqClient MQClient, clusterId, sagaSubmitterId string) (string, string) {
	key := fmt.Sprintf("%s_traxcoord_saga_submitter_%s_outbox", clusterId, sagaSubmitterId)
	return mqClient.GetPublishNodeName(key), mqClient.GetSubscribeNodeName(key)
}

func getControlBusPublishNodeName(mqClient MQClient, clusterId string) string {
	nodeName := mqClient.GetPublishNodeName(
		fmt.Sprintf("%s_trax_coordinators_control_bus", clusterId))
	return nodeName
}

func getControlSubscribeNodeNameByAffinity(mqClient MQClient, clusterId, affinityGroup string) string {
	nodeName := mqClient.GetSubscribeNodeName(
		fmt.Sprintf("%s_trax_coordinator_%s_control_inbox", clusterId, affinityGroup))
	return nodeName
}

func getEventBusNodeNames(mqClient MQClient, clusterId string) (string, string) {
	nodeKey := fmt.Sprintf("%s_trax_event_bus", clusterId)
	publishNodeName := mqClient.GetPublishNodeName(nodeKey)
	subscribeNodeName := mqClient.GetSubscribeNodeName(nodeKey)
	return publishNodeName, subscribeNodeName
}

// --- Topic Exchange naming functions ---

// getStepTopicExchangeName returns the per-cluster topic exchange name for saga step communication.
// Format: x_{clusterId}_trax_saga_steps
func getStepTopicExchangeName(clusterId string) string {
	return fmt.Sprintf("x_%s_trax_saga_steps", clusterId)
}

// getStepRequestRoutingKey generates the routing key for sending
// an execution/compensation request from coordinator to executor.
// Format: {clusterId}.{affinity}.{sagaTemplate}.{stepTemplate}.request
func getStepRequestRoutingKey(clusterId, affinityGroup, sagaTemplateId, sagaStepTemplateId string) string {
	return fmt.Sprintf("%s.%s.%s.%s.request",
		clusterId, affinityGroup, sagaTemplateId, sagaStepTemplateId)
}

// getStepResponseRoutingKey generates the routing key for sending
// an execution result from executor back to coordinator.
// Format: {clusterId}.{affinity}.{sagaTemplate}.{stepTemplate}.response
func getStepResponseRoutingKey(clusterId, affinityGroup, sagaTemplateId, sagaStepTemplateId string) string {
	return fmt.Sprintf("%s.%s.%s.%s.response",
		clusterId, affinityGroup, sagaTemplateId, sagaStepTemplateId)
}

// getExecutorInboxBindingKey generates the binding key pattern for an executor's inbox queue.
// Uses wildcard for affinity since executors serve all affinity groups.
// Format: {clusterId}.*.{sagaTemplate}.{stepTemplate}.request
func getExecutorInboxBindingKey(clusterId, sagaTemplateId, sagaStepTemplateId string) string {
	return fmt.Sprintf("%s.*.%s.%s.request",
		clusterId, sagaTemplateId, sagaStepTemplateId)
}

// getCoordinatorResultsBindingKey generates the binding key pattern for a coordinator's
// aggregated results queue. Receives all step responses for a given affinity.
// Format: {clusterId}.{affinity}.*.*.response
func getCoordinatorResultsBindingKey(clusterId, affinityGroup string) string {
	return fmt.Sprintf("%s.%s.*.*.response",
		clusterId, affinityGroup)
}

// getExecutorInboxQueueName generates the queue name for an executor's inbox.
// One queue per (clusterId, sagaTemplate, stepTemplate) -- shared across all affinities.
func getExecutorInboxQueueName(clusterId, sagaTemplateId, sagaStepTemplateId string) string {
	return fmt.Sprintf("q_%s_trax_executor_%s_%s_inbox",
		clusterId, sagaTemplateId, sagaStepTemplateId)
}

// getCoordinatorResultsQueueName generates the queue name for a coordinator's
// aggregated results queue. One queue per (clusterId, affinity).
func getCoordinatorResultsQueueName(clusterId, affinityGroup string) string {
	return fmt.Sprintf("q_%s_trax_coordinator_%s_results",
		clusterId, affinityGroup)
}
