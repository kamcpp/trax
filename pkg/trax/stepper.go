package trax

import (
	"context"
)

type SagaStepLogicFn func(ctx context.Context) error

type SagaInstanceStepper interface {
	RunStepExecutionLoop(
		ctx context.Context,
		affinity string,
		tenantId, zoneId, sagaId, sagaStepId string,
		logicFn SagaStepLogicFn,
	) error
}

type defaultSagaInstanceStepper struct {
	mqClient         MQClient
	coordinatorProxy CoordinatorProxy
}

func NewDefaultSagaInstanceStepper(
	mqClient MQClient,
	coordinatorProxy CoordinatorProxy,
) SagaInstanceStepper {
	return &defaultSagaInstanceStepper{
		mqClient:         mqClient,
		coordinatorProxy: coordinatorProxy,
	}
}

func (s *defaultSagaInstanceStepper) RunStepExecutionLoop(
	ctx context.Context,
	affinity string,
	tenantId, zoneId, sagaTemplateId, sagaStepTemplateId string,
	logicFn SagaStepLogicFn,
) error {
	// sagaStepTemplate, err := s.coordinatorProxy.GetSagaStepTemplate(
	// 	ctx,
	// 	affinity,
	// 	tenantId, zoneId, sagaTemplateId, sagaStepTemplateId,
	// )
	// if err != nil {
	// 	return err
	// }
	// stepOutboxNodeKey := GenerateSagaStepNodeKeyFromTemplate(
	// 	sagaStepTemplate.OutboxNodeNameTemplate,
	// 	affinity,
	// 	tenantId, zoneId, sagaTemplateId, sagaStepTemplateId,
	// )
	// stepCompensationOutboxNodeKey := GenerateSagaStepNodeKeyFromTemplate(
	// 	sagaStepTemplate.CompensationOutboxNodeNameTemplate,
	// 	affinity,
	// 	tenantId, zoneId, sagaTemplateId, sagaStepTemplateId,
	// )
	// s.mqClient.ConsumeNodeAsync(
	// 	ctx,
	// 	s.mqClient.GetSubscribeNodeName(stepOutboxNodeKey),
	// 	func(ctx context.Context, messageType, contentType string, msg *TraxMessage) error {
	// 		return nil
	// 	},
	// 	func(ctx context.Context, err error) error {
	// 		// handle error
	// 		return nil
	// 	},
	// )
	// s.mqClient.ConsumeNodeAsync(
	// 	ctx,
	// 	s.mqClient.GetSubscribeNodeName(stepCompensationOutboxNodeKey),
	// 	func(ctx context.Context, messageType, contentType string, msg *TraxMessage) error {
	// 		return nil
	// 	},
	// 	func(ctx context.Context, err error) error {
	// 		// handle error
	// 		return nil
	// 	},
	// )
	return nil
}
