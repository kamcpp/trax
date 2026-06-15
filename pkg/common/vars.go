package common

var (
	AgoraApiKey      string
	AgoraAdminApiKey string
)

const (
	MainRedisDB            int = 1
	HistoryRedisDB         int = 6
	BlockchainRedisDB      int = 9
	ExchangeDataRedisDB    int = 10
	TreasuryDataRedisDB    int = 11
	MarketDataIndexRedisDB int = 12
	TradeIndexerRedisDB    int = 13
	TreasuryIndexerRedisDB int = 14
)
