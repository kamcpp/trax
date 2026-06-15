package execplhelpers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/execds"
	"github.com/kamcpp/trax/pkg/execpl"
	exchangemodel "github.com/kamcpp/trax/pkg/marketds/model/exchange"
	mqexchange "github.com/kamcpp/trax/pkg/mq/exchange"
)

func PublishCommandEvent(ctx context.Context, command *execpl.Command, commandEventStatus string) error {
	return nil
}

func PublishOffChainOrderEvent(
	ctx context.Context,
	command *execpl.Command,
	dsOrder *execds.Order,
	orderEventType int,
) {
	internalPublishOffChainOrderEvent(ctx, command, dsOrder, orderEventType, "", "")
}

func PublishOffChainOrderEventWithExecMessage(
	ctx context.Context,
	command *execpl.Command,
	dsOrder *execds.Order,
	orderEventType int,
	execMsgEncoding, execMsg string,
) {
	internalPublishOffChainOrderEvent(
		ctx, command, dsOrder, orderEventType, execMsgEncoding, execMsg)
}

func internalPublishOffChainOrderEvent(
	ctx context.Context,
	command *execpl.Command,
	dsOrder *execds.Order,
	orderEventType int,
	execMsgEncoding, execMsg string,
) {
	exchangemodel.PrepareOrderEventSRecords(
		ctx,
		command, dsOrder,
		nil /* client */, nil /* block */, nil /* tx */, nil, /* log */
		nil /* pair */, nil /* baseTokenInfo */, nil, /* quoteTokenInfo */
		nil /* order */, nil, /* other order */
		orderEventType, "execpl.v1", common.MustDecimalStrToBig("0"), "{}",
		nil /* trade */, nil /* bidOrder */, nil, /* askOrder */
		func(record *exchangemodel.OrderEventSRecord) {
			record.ExecMsgEncoding = execMsgEncoding
			record.ExecMsg = execMsg
			recordBytes, err := json.Marshal(record)
			if err != nil {
				panic(err)
			}
			for {
				err = mqexchange.PublishToExchangeRealtimeExchange(
					ctx, exchangemodel.OrderEventSRecordTypeId, "application/json", recordBytes)
				if err != nil {
					common.L.Error(fmt.Sprintf(
						"queing exchange.order_event__s record failed: %v", err),
						common.F(ctx)...)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		})
	exchangemodel.PrepareOrderEventHRecords(
		ctx,
		command, dsOrder,
		nil /* client */, nil /* block */, nil /* tx */, nil, /* log */
		nil /* pair */, nil /* baseTokenInfo */, nil, /* quoteTokenInfo */
		nil /* order */, nil, /* other order */
		orderEventType, "execpl.v1", common.MustDecimalStrToBig("0"), "{}",
		nil /* trade */, nil /* bidOrder */, nil, /* askOrder */
		func(record *exchangemodel.OrderEventHRecord) {
			record.ExecMsgEncoding = execMsgEncoding
			record.ExecMsg = execMsg
			recordBytes, err := json.Marshal(record)
			if err != nil {
				panic(err)
			}
			for {
				err = mqexchange.PublishToExchangeRealtimeExchange(
					ctx, exchangemodel.OrderEventHRecordTypeId, "application/json", recordBytes)
				if err != nil {
					common.L.Error(fmt.Sprintf(
						"queing exchange.order_event__h record failed: %v", err),
						common.F(ctx)...)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		})
	exchangemodel.PrepareOrderEventSiRecords(
		ctx,
		command, dsOrder,
		nil /* client */, nil /* block */, nil /* tx */, nil, /* log */
		nil /* pair */, nil /* baseTokenInfo */, nil, /* quoteTokenInfo */
		nil /* order */, nil, /* other order */
		orderEventType, "execpl.v1", common.MustDecimalStrToBig("0"), "{}",
		nil /* trade */, nil /* bidOrder */, nil, /* askOrder */
		func(record *exchangemodel.OrderEventSiRecord) {
			record.ExecMsgEncoding = execMsgEncoding
			record.ExecMsg = execMsg
			recordBytes, err := json.Marshal(record)
			if err != nil {
				panic(err)
			}
			for {
				err = mqexchange.PublishToExchangeExchange(
					ctx, exchangemodel.OrderEventSiRecordTypeId, "application/json", recordBytes)
				if err != nil {
					common.L.Error(fmt.Sprintf(
						"queing exchange.order_event__si record failed: %v", err),
						common.F(ctx)...)
					time.Sleep(5 * time.Second)
					continue
				}
				break
			}
		})
}

func PublishExecutionReport(
	ctx context.Context,
	participantOrderId string,
	transactTs int64,
	quantity,
	price,
	remainingQuantity,
	averagePrice float64,
	executionType,
	orderStatus,
	executionTransactionType string,
	data string,
) error {
	os := &execpl.OutboundSignal{
		Type:   string(execpl.OutboundSignalTypeEnum_ExecutionReport),
		Output: map[string]*execpl.Value{},
		Extra:  map[string]string{},
	}
	osout := &os.Output
	participantId, _, _, _, err := common.ParseParticipantOrderId(participantOrderId)
	if err != nil {
		return err
	}
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ParticipantOrderId), participantOrderId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ExecutionId),
		common.SecureRandomString(execpl.ExecutionIdLength))
	// TODO(kam): support multiple currencies
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_Currency), "USD")
	execpl.InsertFloat64Value(
		osout, string(execpl.OutboundSignalOutputKeyEnum_QuantityFloat64), quantity)
	execpl.InsertFloat64Value(
		osout, string(execpl.OutboundSignalOutputKeyEnum_PriceFloat64), price)
	execpl.InsertFloat64Value(
		osout, string(execpl.OutboundSignalOutputKeyEnum_RemainingQuantityFloat64), remainingQuantity)
	execpl.InsertFloat64Value(
		osout, string(execpl.OutboundSignalOutputKeyEnum_AveragePriceFloat64), averagePrice)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ExecutionType), executionType)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_OrderStatus), orderStatus)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ExecutionTransactionType),
		executionTransactionType)
	execpl.InsertIntegerValue(osout, string(execpl.OutboundSignalOutputKeyEnum_TransactTs), transactTs)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_Data), data)
	osBytes, _ := json.Marshal(os)
	return mqexchange.PublishToExchangeOutboundSignalsExchange(
		ctx,
		participantId,
		string(execpl.ExecutionPipelineMessageTypeEnum_OutboundSignal),
		"application/json",
		osBytes,
	)
}

func PublishSecurityDefinition(
	ctx context.Context,
	participantId,
	requestId,
	securityRequestId string,
	totalNumberOfSecurities int64,
	securityIndex int64,
	symbol string,
	decimals int32,
	currency string,
	currencyDecimals int32,
	data string,
	tradeDate string,
	noTickRules int64,
	startTickPriceRange string,
	endTickPriceRange string,
	tickIncrement string,
	securityExchange string,
	securityID string,
	securityIDSource string,
	noSecurityAltID int64,
	securityAltID string,
	securityAltIDSource string,
	securityType string,
	cfiCode string,
	securityDesc string,
	minPriceIncrement string,
	priceType string,
	securityStatus string,
) error {
	os := &execpl.OutboundSignal{
		Type:   string(execpl.OutboundSignalTypeEnum_SecurityDefinition),
		Output: map[string]*execpl.Value{},
		Extra:  map[string]string{},
	}
	osout := &os.Output
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_RequestId), requestId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityRequestId), securityRequestId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ResponseId),
		common.SecureRandomString(execpl.ResponseIdLength))
	execpl.InsertIntegerValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_TotalNumberOfSecurities), totalNumberOfSecurities)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_Symbol), symbol)
	execpl.InsertIntegerValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_Decimals), int64(decimals))
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_Currency), currency)
	execpl.InsertIntegerValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_CurrencyDecimals), int64(currencyDecimals))
	execpl.InsertIntegerValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityIndex), securityIndex)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_Data), data)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_TradeDate), tradeDate)
	execpl.InsertIntegerValue(osout, string(execpl.OutboundSignalOutputKeyEnum_NoTickRules), noTickRules)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_StartTickPriceRange), startTickPriceRange)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_EndTickPriceRange), endTickPriceRange)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_TickIncrement), tickIncrement)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityID), securityID)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityIDSource), securityIDSource)
	execpl.InsertIntegerValue(osout, string(execpl.OutboundSignalOutputKeyEnum_NoSecurityAltID), noSecurityAltID)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityType), securityType)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_CFICode), cfiCode)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityDesc), securityDesc)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_MinPriceIncrement), minPriceIncrement)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_PriceType), priceType)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityStatus), securityStatus)
	if securityExchange != "" {
		execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityExchange), securityExchange)
	}
	if securityAltID != "" {
		execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityAltID), securityAltID)
	}
	if securityAltIDSource != "" {
		execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_SecurityAltIDSource), securityAltIDSource)
	}

	osBytes, _ := json.Marshal(os)
	return mqexchange.PublishToExchangeOutboundSignalsExchange(
		ctx,
		participantId,
		string(execpl.ExecutionPipelineMessageTypeEnum_OutboundSignal),
		"application/json",
		osBytes,
	)
}

func PublishOrderCancelReject(
	ctx context.Context,
	participantOrderId string,
	transactTs int64,
	orderId,
	clientOrderId,
	origClientOrderId,
	orderStatus,
	responseTo,
	reason,
	data string,
) error {
	os := &execpl.OutboundSignal{
		Type:   string(execpl.OutboundSignalTypeEnum_OrderCancelReject),
		Output: map[string]*execpl.Value{},
		Extra:  map[string]string{},
	}
	osout := &os.Output
	participantId, _, _, _, err := common.ParseParticipantOrderId(participantOrderId)
	if err != nil {
		return err
	}
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ParticipantOrderId), participantOrderId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_OrderId), orderId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ClientOrderId), clientOrderId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_OriginalClientOrderId), origClientOrderId)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_OrderStatus), orderStatus)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_ResponseTo), responseTo)
	execpl.InsertStringValue(
		osout, string(execpl.OutboundSignalOutputKeyEnum_Reason), reason)
	execpl.InsertIntegerValue(osout, string(execpl.OutboundSignalOutputKeyEnum_TransactTs), transactTs)
	execpl.InsertStringValue(osout, string(execpl.OutboundSignalOutputKeyEnum_Data), data)
	osBytes, _ := json.Marshal(os)
	return mqexchange.PublishToExchangeOutboundSignalsExchange(
		ctx,
		participantId,
		string(execpl.ExecutionPipelineMessageTypeEnum_OutboundSignal),
		"application/json",
		osBytes,
	)
}
