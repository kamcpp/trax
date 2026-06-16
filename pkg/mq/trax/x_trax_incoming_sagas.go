package mqtrax

import (
	"context"

	mqcommon "github.com/xshyft/trax/pkg/mq/common"
)

const (
	traxIncomingSagasSystemKey = "trax_incoming_sagas"
)

func InitTraxIncomingSagasSystem(ctx context.Context) error {
	return mqcommon.InitExchangeToMultipleQueuesByKey(ctx, traxIncomingSagasSystemKey, []string{})
}

func PublishToTraxIncomingSagasExchange(ctx context.Context, messageType, contentType string, body []byte) error {
	return mqcommon.PublishToExchange(ctx, traxIncomingSagasSystemKey, messageType, contentType, body)
}

func ConsumeTraxIncomingSagasQueueAsync(
	ctx context.Context,
	cb func(ctx context.Context, messageType, contentType string, body []byte) error,
) func() {
	return mqcommon.ConsumeQueueAsync(ctx, traxIncomingSagasSystemKey, cb)
}
