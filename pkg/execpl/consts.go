package execpl

const (
	RequestIdLength   = 36
	CommandIdLength   = 36
	ExecutionIdLength = 36
	ResponseIdLength  = 36
)

// ValueTypeEnum represents the type of a value in execution pipeline
type ValueTypeEnum string

const (
	ValueTypeEnum_Unknown   ValueTypeEnum = "UNKNOWN"
	ValueTypeEnum_Integer   ValueTypeEnum = "VALUE_TYPE_ENUM_INTEGER"
	ValueTypeEnum_Float64   ValueTypeEnum = "VALUE_TYPE_ENUM_FLOAT64"
	ValueTypeEnum_String    ValueTypeEnum = "VALUE_TYPE_ENUM_STRING"
	ValueTypeEnum_HexString ValueTypeEnum = "VALUE_TYPE_ENUM_HEXSTRING"
	ValueTypeEnum_Base64    ValueTypeEnum = "VALUE_TYPE_ENUM_BASE64"
	ValueTypeEnum_Other     ValueTypeEnum = "VALUE_TYPE_ENUM_OTHER"
)

// SideTypeEnum represents the side of an order (buy/sell).
// Values match fin.SideTypeEnum for interoperability.
type SideTypeEnum string

const (
	SideTypeEnum_Unknown          SideTypeEnum = "UNKNOWN"
	SideTypeEnum_Buy              SideTypeEnum = "SIDE_TYPE_ENUM_BUY"
	SideTypeEnum_Sell             SideTypeEnum = "SIDE_TYPE_ENUM_SELL"
	SideTypeEnum_BuyMinus         SideTypeEnum = "SIDE_TYPE_ENUM_BUY_MINUS"
	SideTypeEnum_SellPlus         SideTypeEnum = "SIDE_TYPE_ENUM_SELL_PLUS"
	SideTypeEnum_SellShort        SideTypeEnum = "SIDE_TYPE_ENUM_SELL_SHORT"
	SideTypeEnum_SellShortExempt  SideTypeEnum = "SIDE_TYPE_ENUM_SELL_SHORT_EXEMPT"
	SideTypeEnum_Undisclosed      SideTypeEnum = "SIDE_TYPE_ENUM_UNDISCLOSED"
	SideTypeEnum_Cross            SideTypeEnum = "SIDE_TYPE_ENUM_CROSS"
	SideTypeEnum_CrossShort       SideTypeEnum = "SIDE_TYPE_ENUM_CROSS_SHORT"
	SideTypeEnum_CrossShortExempt SideTypeEnum = "SIDE_TYPE_ENUM_CROSS_SHORT_EXEMPT"
	SideTypeEnum_AsDefined        SideTypeEnum = "SIDE_TYPE_ENUM_AS_DEFINED"
	SideTypeEnum_Opposite         SideTypeEnum = "SIDE_TYPE_ENUM_OPPOSITE"
	SideTypeEnum_Subscribe        SideTypeEnum = "SIDE_TYPE_ENUM_SUBSCRIBE"
	SideTypeEnum_Redeem           SideTypeEnum = "SIDE_TYPE_ENUM_REDEEM"
	SideTypeEnum_Lend             SideTypeEnum = "SIDE_TYPE_ENUM_LEND"
	SideTypeEnum_Borrow           SideTypeEnum = "SIDE_TYPE_ENUM_BORROW"
	SideTypeEnum_Other            SideTypeEnum = "SIDE_TYPE_ENUM_OTHER"
)

// OrderTypeEnum represents the type of an order.
// Values match fin.OrderTypeEnum for interoperability.
type OrderTypeEnum string

const (
	OrderTypeEnum_Unknown OrderTypeEnum = "UNKNOWN"
	OrderTypeEnum_Limit   OrderTypeEnum = "ORDER_TYPE_ENUM_LIMIT"
	OrderTypeEnum_Market  OrderTypeEnum = "ORDER_TYPE_ENUM_MARKET"
	OrderTypeEnum_Other   OrderTypeEnum = "ORDER_TYPE_ENUM_OTHER"
)

// TimeInForceTypeEnum represents the time-in-force for an order.
// Values match fin.TimeInForceEnum for interoperability.
// Note: constant values here use "TYPE_ENUM" prefix for backward compatibility with stored data.
type TimeInForceTypeEnum string

const (
	TimeInForceTypeEnum_Unknown           TimeInForceTypeEnum = "UNKNOWN"
	TimeInForceTypeEnum_Day               TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_DAY"
	TimeInForceTypeEnum_GoodTillCancel    TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_GOOD_TILL_CANCEL"
	TimeInForceTypeEnum_GoodTillCrossing  TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_GOOD_TILL_CROSSING"
	TimeInForceTypeEnum_GoodTillDate      TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_GOOD_TILL_DATE"
	TimeInForceTypeEnum_AtTheOpening      TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_AT_THE_OPENING"
	TimeInForceTypeEnum_FillOrKill        TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_FILL_OR_KILL"
	TimeInForceTypeEnum_ImmediateOrCancel TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_IMMEDIATE_OR_CANCEL"
	TimeInForceTypeEnum_Other             TimeInForceTypeEnum = "TIME_IN_FORCE_TYPE_ENUM_OTHER"
)

// ExecutionPipelineMessageTypeEnum represents the type of execution pipeline message
type ExecutionPipelineMessageTypeEnum string

const (
	ExecutionPipelineMessageTypeEnum_Unknown        ExecutionPipelineMessageTypeEnum = "UNKNOWN"
	ExecutionPipelineMessageTypeEnum_Command        ExecutionPipelineMessageTypeEnum = "EXECUTION_PIPELINE_MESSAGE_TYPE_ENUM_COMMAND"
	ExecutionPipelineMessageTypeEnum_Envelope       ExecutionPipelineMessageTypeEnum = "EXECUTION_PIPELINE_MESSAGE_TYPE_ENUM_ENVELOPE"
	ExecutionPipelineMessageTypeEnum_OutboundSignal ExecutionPipelineMessageTypeEnum = "EXECUTION_PIPELINE_MESSAGE_TYPE_ENUM_OUTBOUND_SIGNAL"
	ExecutionPipelineMessageTypeEnum_Trax           ExecutionPipelineMessageTypeEnum = "EXECUTION_PIPELINE_MESSAGE_TYPE_ENUM_TRAX"
	ExecutionPipelineMessageTypeEnum_Other          ExecutionPipelineMessageTypeEnum = "EXECUTION_PIPELINE_MESSAGE_TYPE_ENUM_OTHER"
)

// CommandOriginEnum represents the origin of a command
type CommandOriginEnum string

const (
	CommandOriginEnum_Unknown CommandOriginEnum = "UNKNOWN"
	CommandOriginEnum_FIX     CommandOriginEnum = "COMMAND_ORIGIN_ENUM_FIX"
	CommandOriginEnum_REST    CommandOriginEnum = "COMMAND_ORIGIN_ENUM_REST"
	CommandOriginEnum_Other   CommandOriginEnum = "COMMAND_ORIGIN_ENUM_OTHER"
)

// CommandMethodEnum represents the HTTP method of a command
type CommandMethodEnum string

const (
	CommandMethodEnum_Unknown CommandMethodEnum = "UNKNOWN"
	CommandMethodEnum_Post    CommandMethodEnum = "COMMAND_METHOD_ENUM_POST"
	CommandMethodEnum_Get     CommandMethodEnum = "COMMAND_METHOD_ENUM_GET"
	CommandMethodEnum_Put     CommandMethodEnum = "COMMAND_METHOD_ENUM_PUT"
	CommandMethodEnum_Delete  CommandMethodEnum = "COMMAND_METHOD_ENUM_DELETE"
	CommandMethodEnum_Other   CommandMethodEnum = "COMMAND_METHOD_ENUM_OTHER"
)

// CommandOperationTypeEnum represents the type of command operation
type CommandOperationTypeEnum string

const (
	CommandOperationTypeEnum_Unknown            CommandOperationTypeEnum = "UNKNOWN"
	CommandOperationTypeEnum_NewOrderSingle     CommandOperationTypeEnum = "COMMAND_OPERATION_TYPE_ENUM_NEW_ORDER_SINGLE"
	CommandOperationTypeEnum_OrderCancelRequest CommandOperationTypeEnum = "COMMAND_OPERATION_TYPE_ENUM_ORDER_CANCEL_REQUEST"
	CommandOperationTypeEnum_Other              CommandOperationTypeEnum = "COMMAND_OPERATION_TYPE_ENUM_OTHER"
)

// OutboundSignalTypeEnum represents the type of outbound signal
type OutboundSignalTypeEnum string

const (
	OutboundSignalTypeEnum_Unknown            OutboundSignalTypeEnum = "UNKNOWN"
	OutboundSignalTypeEnum_ExecutionReport    OutboundSignalTypeEnum = "OUTBOUND_SIGNAL_TYPE_ENUM_EXECUTION_REPORT"
	OutboundSignalTypeEnum_SecurityDefinition OutboundSignalTypeEnum = "OUTBOUND_SIGNAL_TYPE_ENUM_SECURITY_DEFINITION"
	OutboundSignalTypeEnum_OrderCancelReject  OutboundSignalTypeEnum = "OUTBOUND_SIGNAL_TYPE_ENUM_ORDER_CANCEL_REJECT"
	OutboundSignalTypeEnum_Other              OutboundSignalTypeEnum = "OUTBOUND_SIGNAL_TYPE_ENUM_OTHER"
)

// CommandArgumentKeyEnum represents the key for command arguments
type CommandArgumentKeyEnum string

const (
	CommandArgumentKeyEnum_Unknown CommandArgumentKeyEnum = "UNKNOWN"

	CommandArgumentKeyEnum_ParticipantId         CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PARTICIPANT_ID"
	CommandArgumentKeyEnum_ParticipantOrderId    CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PARTICIPANT_ORDER_ID"
	CommandArgumentKeyEnum_ExchangeOrderHash     CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_EXCHANGE_ORDER_HASH"
	CommandArgumentKeyEnum_ParticipantTransactTs CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PARTICIPANT_TRANSACT_TS"
	CommandArgumentKeyEnum_ParticipantAccountId  CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PARTICIPANT_ACCOUNT_ID"
	CommandArgumentKeyEnum_InvestorAccountId     CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_INVESTOR_ACCOUNT_ID"
	CommandArgumentKeyEnum_ExecutorAccountId     CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_EXECUTOR_ACCOUNT_ID"

	CommandArgumentKeyEnum_ClientOrderId         CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_CLIENT_ORDER_ID"
	CommandArgumentKeyEnum_OriginalClientOrderId CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_ORIGINAL_CLIENT_ORDER_ID"

	CommandArgumentKeyEnum_Side            CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_SIDE"
	CommandArgumentKeyEnum_Symbol          CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_SYMBOL"
	CommandArgumentKeyEnum_OrderType       CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_ORDER_TYPE"
	CommandArgumentKeyEnum_Price           CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PRICE"
	CommandArgumentKeyEnum_PriceFloat64    CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PRICE_FLOAT64"
	CommandArgumentKeyEnum_Currency        CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_CURRENCY"
	CommandArgumentKeyEnum_Quantity        CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_QUANTITY"
	CommandArgumentKeyEnum_QuantityFloat64 CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_QUANTITY_FLOAT64"
	CommandArgumentKeyEnum_TimeInforce     CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_TIME_IN_FORCE"
	CommandArgumentKeyEnum_EffectiveTs     CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_EFFECTIVE_TS"
	CommandArgumentKeyEnum_ExpireTs        CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_EXPIRE_TS"
	CommandArgumentKeyEnum_ParticipantData CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_PARTICIPANT_DATA"

	CommandArgumentKeyEnum_Other CommandArgumentKeyEnum = "COMMAND_ARGUMENT_KEY_ENUM_OTHER"
)

// ExecutionTypeEnum represents the type of execution.
// Values match fin.ExecutionTypeEnum for interoperability.
type ExecutionTypeEnum string

const (
	ExecutionTypeEnum_Unknown        ExecutionTypeEnum = "UNKNOWN"
	ExecutionTypeEnum_New            ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_NEW"
	ExecutionTypeEnum_PartialFill    ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_PARTIAL_FILL"
	ExecutionTypeEnum_Fill           ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_FILL"
	ExecutionTypeEnum_DoneForDay     ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_DONE_FOR_DAY"
	ExecutionTypeEnum_Canceled       ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_CANCELED"
	ExecutionTypeEnum_Replaced       ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_REPLACED"
	ExecutionTypeEnum_PendingCancel  ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_PENDING_CANCEL"
	ExecutionTypeEnum_Stopped        ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_STOPPED"
	ExecutionTypeEnum_Rejected       ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_REJECTED"
	ExecutionTypeEnum_Suspended      ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_SUSPENDED"
	ExecutionTypeEnum_PendingNew     ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_PENDING_NEW"
	ExecutionTypeEnum_Calculated     ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_CALCULATED"
	ExecutionTypeEnum_Expired        ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_EXPIRED"
	ExecutionTypeEnum_Restated       ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_RESTATED"
	ExecutionTypeEnum_PendingReplace ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_PENDING_REPLACE"
	ExecutionTypeEnum_Other          ExecutionTypeEnum = "EXECUTION_TYPE_ENUM_OTHER"
)

// OrderStatusEnum represents the status of an order.
// Values match fin.OrderStatusEnum for interoperability.
type OrderStatusEnum string

const (
	OrderStatusEnum_Unknown        OrderStatusEnum = "UNKNOWN"
	OrderStatusEnum_New            OrderStatusEnum = "ORDER_STATUS_ENUM_NEW"
	OrderStatusEnum_PartialFill    OrderStatusEnum = "ORDER_STATUS_ENUM_PARTIAL_FILL"
	OrderStatusEnum_Fill           OrderStatusEnum = "ORDER_STATUS_ENUM_FILL"
	OrderStatusEnum_DoneForDay     OrderStatusEnum = "ORDER_STATUS_ENUM_DONE_FOR_DAY"
	OrderStatusEnum_Canceled       OrderStatusEnum = "ORDER_STATUS_ENUM_CANCELED"
	OrderStatusEnum_Replaced       OrderStatusEnum = "ORDER_STATUS_ENUM_REPLACED"
	OrderStatusEnum_PendingCancel  OrderStatusEnum = "ORDER_STATUS_ENUM_PENDING_CANCEL"
	OrderStatusEnum_Stopped        OrderStatusEnum = "ORDER_STATUS_ENUM_STOPPED"
	OrderStatusEnum_Rejected       OrderStatusEnum = "ORDER_STATUS_ENUM_REJECTED"
	OrderStatusEnum_Suspended      OrderStatusEnum = "ORDER_STATUS_ENUM_SUSPENDED"
	OrderStatusEnum_PendingNew     OrderStatusEnum = "ORDER_STATUS_ENUM_PENDING_NEW"
	OrderStatusEnum_Calculated     OrderStatusEnum = "ORDER_STATUS_ENUM_CALCULATED"
	OrderStatusEnum_Expired        OrderStatusEnum = "ORDER_STATUS_ENUM_EXPIRED"
	OrderStatusEnum_Restated       OrderStatusEnum = "ORDER_STATUS_ENUM_RESTATED"
	OrderStatusEnum_PendingReplace OrderStatusEnum = "ORDER_STATUS_ENUM_PENDING_REPLACE"
	OrderStatusEnum_Other          OrderStatusEnum = "ORDER_STATUS_ENUM_OTHER"
)

// ExecutionTransactionTypeEnum represents the type of execution transaction
type ExecutionTransactionTypeEnum string

const (
	ExecutionTransactionTypeEnum_Unknown ExecutionTransactionTypeEnum = "UNKNOWN"
	ExecutionTransactionTypeEnum_New     ExecutionTransactionTypeEnum = "EXECUTION_TRANSACTION_TYPE_ENUM_NEW"
	ExecutionTransactionTypeEnum_Cancel  ExecutionTransactionTypeEnum = "EXECUTION_TRANSACTION_TYPE_ENUM_CANCEL"
	ExecutionTransactionTypeEnum_Correct ExecutionTransactionTypeEnum = "EXECUTION_TRANSACTION_TYPE_ENUM_CORRECT"
	ExecutionTransactionTypeEnum_Status  ExecutionTransactionTypeEnum = "EXECUTION_TRANSACTION_TYPE_ENUM_STATUS"
	ExecutionTransactionTypeEnum_Other   ExecutionTransactionTypeEnum = "EXECUTION_TRANSACTION_TYPE_ENUM_OTHER"
)

// OutboundSignalOutputKeyEnum represents the key for outbound signal outputs
type OutboundSignalOutputKeyEnum string

const (
	OutboundSignalOutputKeyEnum_Unknown OutboundSignalOutputKeyEnum = "UNKNOWN"

	OutboundSignalOutputKeyEnum_ExecutionId               OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_EXECUTION_ID"
	OutboundSignalOutputKeyEnum_RequestId                 OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_REQUEST_ID"
	OutboundSignalOutputKeyEnum_ResponseId                OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_RESPONSE_ID"
	OutboundSignalOutputKeyEnum_ParticipantOrderId        OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_PARTICIPANT_ORDER_ID"
	OutboundSignalOutputKeyEnum_SecurityRequestId         OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_REQUEST_ID"
	OutboundSignalOutputKeyEnum_OrderId                   OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_ORDER_ID"
	OutboundSignalOutputKeyEnum_ClientOrderId             OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CLIENT_ORDER_ID"
	OutboundSignalOutputKeyEnum_OriginalClientOrderId     OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_ORIGINAL_CLIENT_ORDER_ID"
	OutboundSignalOutputKeyEnum_ExchangeOrderHash         OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_EXCHANGE_ORDER_HASH"
	OutboundSignalOutputKeyEnum_ExecutionType             OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_EXECUTION_TYPE"
	OutboundSignalOutputKeyEnum_OrderStatus               OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_ORDER_STATUS"
	OutboundSignalOutputKeyEnum_ExecutionTransactionType  OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_EXECUTION_TRANSACTION_TYPE"
	OutboundSignalOutputKeyEnum_Symbol                    OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SYMBOL"
	OutboundSignalOutputKeyEnum_Decimals                  OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_DECIMALS"
	OutboundSignalOutputKeyEnum_Side                      OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SIDE"
	OutboundSignalOutputKeyEnum_AveragePrice              OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_AVERAGE_PRICE"
	OutboundSignalOutputKeyEnum_AveragePriceFloat64       OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_AVERAGE_PRICE_FLOAT64"
	OutboundSignalOutputKeyEnum_Price                     OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_PRICE"
	OutboundSignalOutputKeyEnum_PriceFloat64              OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_PRICE_FLOAT64"
	OutboundSignalOutputKeyEnum_Currency                  OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CURRENCY"
	OutboundSignalOutputKeyEnum_CurrencyDecimals          OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CURRENCY_DECIMALS"
	OutboundSignalOutputKeyEnum_Quantity                  OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_QUANTITY"
	OutboundSignalOutputKeyEnum_QuantityFloat64           OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_QUANTITY_FLOAT64"
	OutboundSignalOutputKeyEnum_RemainingQuantity         OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_REMAINING_QUANTITY"
	OutboundSignalOutputKeyEnum_RemainingQuantityFloat64  OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_REMAINING_QUANTITY_FLOAT64"
	OutboundSignalOutputKeyEnum_CumulativeQuantity        OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CUMULATIVE_QUANTITY"
	OutboundSignalOutputKeyEnum_CumulativeQuantityFloat64 OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CUMULATIVE_QUANTITY_FLOAT64"
	OutboundSignalOutputKeyEnum_TransactTs                OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_TRANSACT_TS"
	OutboundSignalOutputKeyEnum_TotalNumberOfSecurities   OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_TOTAL_NUMBER_OF_SECURITIES"
	OutboundSignalOutputKeyEnum_SecurityIndex             OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_INDEX"
	OutboundSignalOutputKeyEnum_Data                      OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_DATA"
	OutboundSignalOutputKeyEnum_Reason                    OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_REASON"
	OutboundSignalOutputKeyEnum_ResponseTo                OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_RESPONSE_TO"

	// Trapets Security Definition fields
	OutboundSignalOutputKeyEnum_TradeDate           OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_TRADE_DATE"
	OutboundSignalOutputKeyEnum_NoTickRules         OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_NO_TICK_RULES"
	OutboundSignalOutputKeyEnum_StartTickPriceRange OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_START_TICK_PRICE_RANGE"
	OutboundSignalOutputKeyEnum_EndTickPriceRange   OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_END_TICK_PRICE_RANGE"
	OutboundSignalOutputKeyEnum_TickIncrement       OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_TICK_INCREMENT"
	OutboundSignalOutputKeyEnum_SecurityExchange    OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_EXCHANGE"
	OutboundSignalOutputKeyEnum_SecurityID          OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_ID"
	OutboundSignalOutputKeyEnum_SecurityIDSource    OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_ID_SOURCE"
	OutboundSignalOutputKeyEnum_NoSecurityAltID     OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_NO_SECURITY_ALT_ID"
	OutboundSignalOutputKeyEnum_SecurityAltID       OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_ALT_ID"
	OutboundSignalOutputKeyEnum_SecurityAltIDSource OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_ALT_ID_SOURCE"
	OutboundSignalOutputKeyEnum_SecurityType        OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_TYPE"
	OutboundSignalOutputKeyEnum_CFICode             OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_CFI_CODE"
	OutboundSignalOutputKeyEnum_SecurityDesc        OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_DESC"
	OutboundSignalOutputKeyEnum_MinPriceIncrement   OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_MIN_PRICE_INCREMENT"
	OutboundSignalOutputKeyEnum_PriceType           OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_PRICE_TYPE"
	OutboundSignalOutputKeyEnum_SecurityStatus      OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_SECURITY_STATUS"

	OutboundSignalOutputKeyEnum_Other OutboundSignalOutputKeyEnum = "OUTBOUND_SIGNAL_OUTPUT_KEY_ENUM_OTHER"
)

// CommandEventStatusTypeEnum represents the status type of a command event
type CommandEventStatusTypeEnum string

const (
	CommandEventStatusTypeEnum_Unknown CommandEventStatusTypeEnum = "UNKNOWN"

	// for all commands
	CommandEventStatusTypeEnum_ReceivedFromOrigin    CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_RECEIVED_FROM_ORIGIN"
	CommandEventStatusTypeEnum_Authorizing           CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_AUTHORIZING"
	CommandEventStatusTypeEnum_Authorized            CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_AUTHORIZED"
	CommandEventStatusTypeEnum_QueueingForProcessing CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_QUEUEING_FOR_PROCESSING"
	CommandEventStatusTypeEnum_QueuedForProcessing   CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_QUEUED_FOR_PROCESSING"
	CommandEventStatusTypeEnum_PickedUpForProcessing CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_PICKED_UP_FOR_PROCESSING"

	// for query commands
	CommandEventStatusTypeEnum_ProducingResults CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_PRODUCING_RESULTS"
	CommandEventStatusTypeEnum_ResultsProduced  CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_RESULTS_PRODUCED"
	CommandEventStatusTypeEnum_QueueingResults  CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_QUEUEING_RESULTS"
	CommandEventStatusTypeEnum_ResultsQueued    CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_RESULTS_QUEUED"
	CommandEventStatusTypeEnum_SendingResults   CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_SENDING_RESULTS"
	CommandEventStatusTypeEnum_ResultsSent      CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_RESULTS_SENT"

	// for mutating commands
	CommandEventStatusTypeEnum_PreparingEnvelope       CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_PREPARING_ENVELOPE"
	CommandEventStatusTypeEnum_EnvelopePrepared        CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_ENVELOPE_PREPARED"
	CommandEventStatusTypeEnum_QueueingForBroadcasting CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_QUEUEING_FOR_BROADCASTING"
	CommandEventStatusTypeEnum_QueuedForBroadcasting   CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_QUEUED_FOR_BROADCASTING"
	CommandEventStatusTypeEnum_CommandProcessed        CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_COMMAND_PROCESSED"
	CommandEventStatusTypeEnum_PickedUpForBroadcasting CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_PICKED_UP_FOR_BROADCASTING"
	CommandEventStatusTypeEnum_Broadcasting            CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_BROADCASTING"
	CommandEventStatusTypeEnum_Broadcasted             CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_BROADCASTED"

	// final states for all commands
	CommandEventStatusTypeEnum_NotAuthorized      CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_NOT_AUTHORIZED"
	CommandEventStatusTypeEnum_CompletedWithError CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_COMPLETED_WITH_ERROR"
	CommandEventStatusTypeEnum_Completed          CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_COMPLETED"

	CommandEventStatusTypeEnum_Other CommandEventStatusTypeEnum = "COMMAND_EVENT_STATUS_TYPE_ENUM_OTHER"
)

// CxlRejResponseToEnum represents the response type for cancel rejection
type CxlRejResponseToEnum string

const (
	CxlRejResponseToEnum_Unknown                   CxlRejResponseToEnum = "UNKNOWN"
	CxlRejResponseToEnum_OrderCancelRequest        CxlRejResponseToEnum = "CXL_REJ_RESPONSE_TO_ENUM_ORDER_CANCEL_REQUEST"
	CxlRejResponseToEnum_OrderCancelReplaceRequest CxlRejResponseToEnum = "CXL_REJ_RESPONSE_TO_ENUM_ORDER_CANCEL_REPLACE_REQUEST"
	CxlRejResponseToEnum_Other                     CxlRejResponseToEnum = "CXL_REJ_RESPONSE_TO_ENUM_OTHER"
)

// CxlRejReasonEnum represents the reason for cancel rejection
type CxlRejReasonEnum string

const (
	CxlRejReasonEnum_Unknown                CxlRejReasonEnum = "UNKNOWN"
	CxlRejReasonEnum_TooLateToCancel        CxlRejReasonEnum = "CXL_REJ_REASON_ENUM_TOO_LATE_TO_CANCEL"
	CxlRejReasonEnum_UnknownOrder           CxlRejReasonEnum = "CXL_REJ_REASON_ENUM_UNKNOWN_ORDER"
	CxlRejReasonEnum_BrokerOption           CxlRejReasonEnum = "CXL_REJ_REASON_ENUM_BROKER_OPTION"
	CxlRejReasonEnum_PendingCancelOrReplace CxlRejReasonEnum = "CXL_REJ_REASON_ENUM_PENDING_CANCEL_OR_REPLACE"
	CxlRejReasonEnum_Other                  CxlRejReasonEnum = "CXL_REJ_REASON_ENUM_OTHER"
)
