package common

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
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

func BigIntToStr(num *big.Int) string {
	numStr := num.String()
	if numStr == "0" {
		return ""
	}
	return numStr
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

// GetServiceBaseURL returns the base URL for a standalone TRAX HTTP service.
// Supported names are limited to the daemons this repo actually runs and tests.
func GetServiceBaseURL(serviceName string) string {
	var envVarName string
	switch strings.ToLower(serviceName) {
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
