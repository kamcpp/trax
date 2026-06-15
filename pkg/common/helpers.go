package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func Paginate[T any](items []T, page, pageSize int) []T {
	skip := (page - 1) * pageSize
	if skip > len(items) {
		skip = len(items)
	}
	end := skip + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[skip:end]
}

func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func AddressToStr(addr ethcommon.Address) string {
	zeroAddressBytes := ethcommon.FromHex("0x0000000000000000000000000000000000000000")
	addrBytes := addr.Bytes()
	if reflect.DeepEqual(addrBytes, zeroAddressBytes) {
		return ""
	}
	return strings.ToLower(addr.String())
}

func BigIntToStr(num *big.Int) string {
	numStr := num.String()
	if numStr == "0" {
		return ""
	}
	return numStr
}

func GetTxSender(tx *types.Transaction) (*ethcommon.Address, error) {
	sender, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
	if err != nil {
		return nil, err
	}
	return &sender, nil
}

func MustStrToInt64(decimalStr DecimalStr) int64 {
	n, err := strconv.ParseInt(string(decimalStr), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("error converting '%s' to int64: %s", string(decimalStr), err.Error()))
	}
	return n
}

func MustStrToUint64(decimalStr DecimalStr) uint64 {
	n, err := strconv.ParseUint(string(decimalStr), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("error converting '%s' to uint64: %s", string(decimalStr), err.Error()))
	}
	return n
}

func MustStrToFloat64(float64Str string) float64 {
	f, err := strconv.ParseFloat(float64Str, 64)
	if err != nil {
		panic(fmt.Sprintf("error converting '%s' to float64: %s", float64Str, err.Error()))
	}
	return f
}

func MustHexStrToBytes32(hexStrWith0x string) [32]byte {
	var buff [32]byte
	{
		decodedBuff, err := hex.DecodeString(hexStrWith0x[2:])
		if err != nil {
			panic(err)
		}
		copy(buff[:], decodedBuff)
	}
	return buff
}

func MustHexStrToBig(hexStr string) *big.Int {
	result := new(big.Int)
	trimmedHexStr := strings.TrimPrefix(hexStr, "0x")
	result, ok := result.SetString(trimmedHexStr, 16)
	if !ok {
		panic(fmt.Sprintf("could not convert hex string to bigint: %s", hexStr))
	}
	return result
}

func MustDecimalStrToBig(decimalStr string) *big.Int {
	result := new(big.Int)
	result, ok := result.SetString(decimalStr, 10)
	if !ok {
		panic(fmt.Sprintf("could not convert decimal string to bigint: %s", decimalStr))
	}
	return result
}

func IsInteger(f float64) bool {
	return f == math.Trunc(f)
}

func SanitizeHexStr(hexStr string) string {
	hexStr = strings.ToLower(hexStr)
	if hexStr == "" || hexStr == "0x" || hexStr == "0x0" || hexStr == "0" {
		return ""
	}
	return strings.TrimPrefix(hexStr, "0x")
}

// IsEVMAccountAddress checks if the given string is a valid EVM account address.
// A valid EVM address has the format: 0x followed by 40 hexadecimal characters.
func IsEVMAccountAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Check if remaining characters are valid hex
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

func ToMilliSecTimestamp(timestamp string) string {
	if len(timestamp) == 0 || timestamp == "0" {
		return "0"
	}
	return timestamp + "000"
}

func MustMarshalToJsonString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("error marshaling to json string: %s [%+v]", err.Error(), v))
	}
	return string(data)
}

// GetServiceBaseURL returns the base URL for an HTTP service from environment variables.
// serviceName: short service name (e.g., "accmgr", "instrmgr", "lcmgr", "lasersvc", "treassvc")
// Automatically maps to full env var names and appends _BASE_URL
// Returns: full base URL including protocol and /api/v1 (e.g., "http://host:port/api/v1")
// Panics if environment variable is not set
func GetServiceBaseURL(serviceName string) string {
	// Map short service names to full environment variable names
	var envVarName string
	switch strings.ToLower(serviceName) {
	case "accmgr":
		envVarName = "ACCOUNT_MANAGER_BASE_URL"
	case "instrmgr":
		envVarName = "INSTRUMENT_MANAGER_BASE_URL"
	case "lasersvc":
		envVarName = "LASER_SERVICE_BASE_URL"
	case "treassvc":
		envVarName = "TREASURY_SERVICE_BASE_URL"
	case "lcmgr":
		envVarName = "ETH_SMART_CONTRACT_MANAGER_BASE_URL"
	case "csdmsggw":
		envVarName = "CSD_MESSAGE_GATEWAY_BASE_URL"
	case "csdsender":
		envVarName = "CSD_SENDER_BASE_URL"
	case "csdrecv":
		envVarName = "CSD_RECEIVER_BASE_URL"
	case "traxcoord", "traxcoord1", "traxcoord2", "traxcoord3":
		envVarName = "TRAX_COORDINATOR_BASE_URL"
	case "test.traxcoord1":
		envVarName = "TRAX_COORDINATOR1_BASE_URL"
	case "test.traxcoord2":
		envVarName = "TRAX_COORDINATOR2_BASE_URL"
	case "test.traxcoord3":
		envVarName = "TRAX_COORDINATOR3_BASE_URL"
	case "traxctrl":
		envVarName = "TRAX_CONTROLLER_BASE_URL"
	case "exchange-traxctrl":
		envVarName = "EXCHANGE_TRAX_CONTROLLER_BASE_URL"
	case "exchange-accmgr":
		envVarName = "EXCHANGE_ACCOUNT_MANAGER_BASE_URL"
	case "exchange-instrmgr":
		envVarName = "EXCHANGE_INSTRUMENT_MANAGER_BASE_URL"
	case "exchange-sdmgr":
		envVarName = "EXCHANGE_SECURITY_DEPOSITORY_MANAGER_BASE_URL"
	case "exchange-configmgr":
		envVarName = "EXCHANGE_CONFIG_MANAGER_BASE_URL"
	case "exchange-listingmgr":
		envVarName = "EXCHANGE_LISTING_MANAGER_BASE_URL"
	case "exchange-tradeidxer":
		envVarName = "EXCHANGE_TRADE_INDEXER_BASE_URL"
	case "exchange-fixreceiver":
		envVarName = "EXCHANGE_FIX_RECEIVER_BASE_URL"
	case "exchange-csdmsggw":
		envVarName = "EXCHANGE_CSD_MESSAGE_GATEWAY_BASE_URL"
	case "prtagent-traxctrl":
		envVarName = "PRTAGENT_TRAX_CONTROLLER_BASE_URL"
	case "prtagent-accmgr":
		envVarName = "PRTAGENT_ACCOUNT_MANAGER_BASE_URL"
	case "prtagent-instrmgr":
		envVarName = "PRTAGENT_INSTRUMENT_MANAGER_BASE_URL"
	case "prtagent-sdmgr":
		envVarName = "PRTAGENT_SECURITY_DEPOSITORY_MANAGER_BASE_URL"
	case "prtagent-configmgr":
		envVarName = "PRTAGENT_CONFIG_MANAGER_BASE_URL"
	case "prtagent-marketmgr":
		envVarName = "PRTAGENT_MARKET_MANAGER_BASE_URL"
	case "prtagent-fixclient":
		envVarName = "PRTAGENT_FIX_CLIENT_BASE_URL"
	case "prtagent-treassvc":
		envVarName = "PRTAGENT_TREASURY_SERVICE_BASE_URL"
	case "prtagent-treasidxer":
		envVarName = "PRTAGENT_TREASURY_INDEXER_BASE_URL"
	case "prtagent-prtagent":
		envVarName = "PRTAGENT_PARTICIPANT_AGENT_GRPC_URL"
	case "iso20022-processor", "iso20022processor":
		envVarName = "ISO20022_PROCESSOR_BASE_URL"
	case "sdmgr":
		envVarName = "SECURITY_DEPOSITORY_MANAGER_BASE_URL"
	case "marketmgr":
		envVarName = "MARKET_MANAGER_BASE_URL"
	case "configmgr":
		envVarName = "CONFIG_MANAGER_BASE_URL"
	case "listingmgr":
		envVarName = "LISTING_MANAGER_BASE_URL"
	case "fixclient":
		envVarName = "FIX_CLIENT_BASE_URL"
	case "signersvc":
		envVarName = "SIGNER_SERVICE_BASE_URL"
	case "tradeidxer":
		envVarName = "TRADE_INDEXER_BASE_URL"
	case "treasidxer":
		envVarName = "TREASURY_INDEXER_BASE_URL"
	case "csd_msggw":
		envVarName = "CSD_MSG_GW_BASE_URL"
	case "csd_accmgr":
		envVarName = "CSD_ACCOUNT_MANAGER_BASE_URL"
	case "actusvc":
		envVarName = "STATE_ACTUATOR_SERVICE_BASE_URL"
	case "prtagent", "prtagentgrpc", "prtagentgrpcsvc":
		envVarName = "PARTICIPANT_AGENT_GRPC_URL"
	case "prtagent-http":
		// Daemon's HTTP testing port (health, clear-caches, setdbname).
		// Distinct from the gRPC URL above because setServiceDatabase
		// drives the registered /api/v1/experimental/testing/setdbname
		// route, which only the HTTP server exposes.
		envVarName = "PRTAGENT_PARTICIPANT_AGENT_HTTP_URL"
	default:
		panic(fmt.Sprintf("unknown service name: %s", serviceName))
	}

	baseURL := os.Getenv(envVarName)
	if baseURL == "" {
		panic(fmt.Sprintf("%s environment variable is not set", envVarName))
	}

	return baseURL
}

// AccountUidPrefix returns the external-grade account_uid prefix for the
// daemon's TRAX cluster. The middle token marks the originating namespace
// so a uid keeps its provenance once it leaves the local DB:
//
//	CSD       → "agora_sd_accuid_"   (security depository)
//	EXCHANGE  → "agora_exch_accuid_"
//	PRTAGENT  → "agora_prta_accuid_"
//
// Reads TRAX_CLUSTER_ID; unset/unknown falls back to sd to preserve the
// pre-existing identifier shape for tooling that hasn't been wired up yet.
func AccountUidPrefix() string {
	switch os.Getenv("TRAX_CLUSTER_ID") {
	case "EXCHANGE":
		return "agora_exch_accuid_"
	case "PRTAGENT":
		return "agora_prta_accuid_"
	default:
		return "agora_sd_accuid_"
	}
}

// GetOptionalServiceBaseURL is like GetServiceBaseURL but returns empty string
// instead of panicking when the environment variable is not set.
// Use this for optional cross-namespace service dependencies (e.g., CSD services
// that may not be configured in all deployments).
func GetOptionalServiceBaseURL(serviceName string) string {
	var envVarName string
	switch strings.ToLower(serviceName) {
	case "csdmsggw":
		envVarName = "CSD_MESSAGE_GATEWAY_BASE_URL"
	case "csd_accmgr":
		envVarName = "CSD_ACCOUNT_MANAGER_BASE_URL"
	case "tradeidxer":
		envVarName = "TRADE_INDEXER_BASE_URL"
	case "treasidxer":
		// Optional in prtagent — only the new GetInvestorTreasuryActivities
		// path needs it. Other RPCs degrade independently.
		envVarName = "TREASURY_INDEXER_BASE_URL"
	case "marketmgr":
		envVarName = "MARKET_MANAGER_BASE_URL"
	default:
		// For non-optional services, delegate to GetServiceBaseURL which panics on missing
		return GetServiceBaseURL(serviceName)
	}

	return os.Getenv(envVarName)
}

// GetLocalizedString retrieves the best matching localized string from a map[string]string.
// Logic:
// 1. If the map has only one entry, return that entry's value
// 2. If currentLocale is provided and exists in the map, return that value
// 3. Otherwise, try the fallback locale "en-US"
// 4. If nothing matches, return the fallbackValue (typically an IID or empty string)
//
// Parameters:
//   - localizedMap: map of locale codes (e.g., "en-US", "fa-IR") to localized strings
//   - currentLocale: the current locale in "xx-XX" format (can be empty)
//   - fallbackValue: value to return if no match found (typically entity IID)
//
// Example usage:
//
//	displayName := GetLocalizedString(entity.DisplayNames, "fa-IR", entity.IID)
//	description := GetLocalizedString(entity.Descriptions, currentLocale, "")
func GetLocalizedString(localizedMap map[string]string, currentLocale string, fallbackValue string) string {
	// If map is empty, return fallback
	if len(localizedMap) == 0 {
		return fallbackValue
	}

	// If only one entry, return it regardless of locale
	if len(localizedMap) == 1 {
		for _, value := range localizedMap {
			return value
		}
	}

	// Try current locale if provided
	if currentLocale != "" {
		if value, ok := localizedMap[currentLocale]; ok {
			return value
		}
	}

	// Try fallback locale "en-US"
	if value, ok := localizedMap["en-US"]; ok {
		return value
	}

	// Return the fallback value (typically IID)
	return fallbackValue
}

// GetBestDescription returns the best available description for an entity.
// This handles both the standard Description field (single string) and the
// Descriptions map (localized strings).
//
// Logic:
// 1. If standard Description field is non-empty, use it
// 2. Otherwise, use GetLocalizedString on the Descriptions map
//
// Parameters:
//   - description: the standard Description field (single string)
//   - descriptions: map of locale codes to localized descriptions
//   - currentLocale: the current locale in "xx-XX" format (can be empty)
//   - fallbackValue: value to return if no description found (typically entity IID)
//
// Example usage:
//
//	desc := GetBestDescription(entity.Description, entity.Descriptions, "fa-IR", entity.IID)
func GetBestDescription(description string, descriptions map[string]string, currentLocale string, fallbackValue string) string {
	// If standard Description field is set, use it
	if description != "" {
		return description
	}

	// Otherwise, use localized descriptions map
	return GetLocalizedString(descriptions, currentLocale, fallbackValue)
}
