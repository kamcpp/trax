package common

type ZonedIdentifier struct {
	Zone             string   `json:"zone"`
	StandardOrFormat string   `json:"standard_or_format"`
	Identifiers      []string `json:"identifiers"`

	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type TextEncodingEnum string

const (
	TextEncodingEnum_Unknown  TextEncodingEnum = "UNKNOWN"
	TextEncodingEnum_ASCII    TextEncodingEnum = "TEXT_ENCODING_ENUM_ASCII"
	TextEncodingEnum_UTF8     TextEncodingEnum = "TEXT_ENCODING_ENUM_UTF8"
	TextEncodingEnum_ISO88591 TextEncodingEnum = "TEXT_ENCODING_ENUM_ISO88591"
)

type StringValue struct {
	Value           string `json:"value"`
	CaseInsensitive bool   `json:"case_insensitive"`
	Encoding        string `json:"encoding"`
}

type AssetTypeEnum string

const (
	AssetTypeEnum_Unknown              AssetTypeEnum = "UNKNOWN"
	AssetTypeEnum_Cash                 AssetTypeEnum = "ASSET_TYPE_ENUM_CASH"
	AssetTypeEnum_OnChain              AssetTypeEnum = "ASSET_TYPE_ENUM_ON_CHAIN"
	AssetTypeEnum_RWA                  AssetTypeEnum = "ASSET_TYPE_ENUM_RWA"
	AssetTypeEnum_Equity               AssetTypeEnum = "ASSET_TYPE_ENUM_EQUITY"
	AssetTypeEnum_FixedIncome          AssetTypeEnum = "ASSET_TYPE_ENUM_FIXED_INCOME"
	AssetTypeEnum_Derivative           AssetTypeEnum = "ASSET_TYPE_ENUM_DERIVATIVE"
	AssetTypeEnum_Commodity            AssetTypeEnum = "ASSET_TYPE_ENUM_COMMODITY"
	AssetTypeEnum_Currency             AssetTypeEnum = "ASSET_TYPE_ENUM_CURRENCY"
	AssetTypeEnum_MutualFund           AssetTypeEnum = "ASSET_TYPE_ENUM_MUTUAL_FUND"
	AssetTypeEnum_ETF                  AssetTypeEnum = "ASSET_TYPE_ENUM_ETF"
	AssetTypeEnum_Index                AssetTypeEnum = "ASSET_TYPE_ENUM_INDEX"
	AssetTypeEnum_RealEstate           AssetTypeEnum = "ASSET_TYPE_ENUM_REAL_ESTATE"
	AssetTypeEnum_Crypto               AssetTypeEnum = "ASSET_TYPE_ENUM_CRYPTO"
	AssetTypeEnum_Stablecoin           AssetTypeEnum = "ASSET_TYPE_ENUM_STABLECOIN"
	AssetTypeEnum_NFT                  AssetTypeEnum = "ASSET_TYPE_ENUM_NFT"
	AssetTypeEnum_Stock                AssetTypeEnum = "ASSET_TYPE_ENUM_STOCK"
	AssetTypeEnum_Bond                 AssetTypeEnum = "ASSET_TYPE_ENUM_BOND"
	AssetTypeEnum_Future               AssetTypeEnum = "ASSET_TYPE_ENUM_FUTURE"
	AssetTypeEnum_Option               AssetTypeEnum = "ASSET_TYPE_ENUM_OPTION"
	AssetTypeEnum_Swap                 AssetTypeEnum = "ASSET_TYPE_ENUM_SWAP"
	AssetTypeEnum_Forward              AssetTypeEnum = "ASSET_TYPE_ENUM_FORWARD"
	AssetTypeEnum_Perpetual            AssetTypeEnum = "ASSET_TYPE_ENUM_PERPETUAL"
	AssetTypeEnum_CFD                  AssetTypeEnum = "ASSET_TYPE_ENUM_CFD"
	AssetTypeEnum_IntellectualProperty AssetTypeEnum = "ASSET_TYPE_ENUM_INTELLECTUAL_PROPERTY"
	AssetTypeEnum_Collectible          AssetTypeEnum = "ASSET_TYPE_ENUM_COLLECTIBLE"
	AssetTypeEnum_Art                  AssetTypeEnum = "ASSET_TYPE_ENUM_ART"
	AssetTypeEnum_VentureCapital       AssetTypeEnum = "ASSET_TYPE_ENUM_VENTURE_CAPITAL"
	AssetTypeEnum_PrivateEquity        AssetTypeEnum = "ASSET_TYPE_ENUM_PRIVATE_EQUITY"
	AssetTypeEnum_Other                AssetTypeEnum = "ASSET_TYPE_ENUM_OTHER"
)

type ZonedSymbol struct {
	Zone             string        `json:"zone"`
	StandardOrFormat string        `json:"standard_or_format"`
	Symbols          []StringValue `json:"symbols"`

	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type ZonedAssetType struct {
	Zone  string          `json:"zone"`
	Types []AssetTypeEnum `json:"types"`

	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

type Asset struct {
	// this is the internal asset id. this can be the most famous asset's identifier
	// e.g., for USD, it can be "USD" or any other internal unique id but it SHOULD NOT
	// be the asset identifier like ISIN, CUSIP, SEDOL, FIGI, TICKER, etc.
	// those should go into ZonedIdentifiers
	Id string `json:"id"`

	Name        string `json:"name"`
	DisplayName string `json:"display_name"`

	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`

	ZonedIdentifiers []ZonedIdentifier `json:"zoned_identifiers"`
	ZonedSymbols     []ZonedSymbol     `json:"zoned_symbols"`
	ZonedAssetTypes  []ZonedAssetType  `json:"zoned_asset_types"`

	Decimals int32 `json:"decimals"`
	Disabled bool  `json:"disabled"`
}
