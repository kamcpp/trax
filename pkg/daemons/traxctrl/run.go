package traxctrl

import (
	"context"

	mqtrax "github.com/kamcpp/trax/pkg/mq/trax"
)

func RunCtrl(ctx context.Context) {

	mqtrax.ConsumeTraxIncomingSagasQueueAsync(ctx, func(ctx context.Context, messageType, contentType string, body []byte) error {
		// if messageType == exchangemodel.TradeSRecordTypeId {
		// 	var record exchangemodel.TradeSRecord
		// 	err := json.Unmarshal(body, &record)
		// 	if err != nil {
		// 		common.L.Error(
		// 			fmt.Sprintf("parsing exchange.trade__s record failed: %v", err), common.F(ctx)...)
		// 		return err
		// 	}
		// 	// fmt.Printf("!!! S %s \n", record.TradeHash)
		// 	err = exchange.InsertTradeSRecord(&record)
		// 	if err != nil {
		// 		common.L.Error(
		// 			fmt.Sprintf("inserting exchange.trade__s record failed: %v", err), common.F(ctx)...)
		// 		return err
		// 	}
		// 	// common.L.Debug("exchange.trade__s inserted successfully.", common.F(ctx)...)
		// 	return nil
		// }
		return nil
	})
}
