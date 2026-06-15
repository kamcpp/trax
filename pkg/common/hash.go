package common

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strings"
)

func GetOnChainOrderHash(
	chainId DecimalStr,
	chainName string,
	engineAddr HexStrWith0x,
	pairId HexStrWith0x,
	orderId DecimalStr,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"agora-onchain-order|%s|%s|%s|%s|%s",
		strings.ToLower(string(chainId)),
		strings.ToLower(chainName),
		strings.ToLower(string(engineAddr)),
		strings.ToLower(string(pairId)),
		strings.ToLower(string(orderId)),
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	hashValue := HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
	// fmt.Printf(">> str: %s, hash: %s\n", str, hashValue)
	return hashValue
}

func GetOffChainOrderHash(
	participantOrderId string,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"agora-offchain-order|%s",
		participantOrderId,
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	hashValue := HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
	// fmt.Printf(">> str: %s, hash: %s\n", str, hashValue)
	return hashValue
}

func GetTradeHash(
	chainId DecimalStr,
	chainName string,
	engineAddr HexStrWith0x,
	pairId HexStrWith0x,
	tradeId DecimalStr,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"agora-trade|%s|%s|%s|%s|%s",
		strings.ToLower(string(chainId)),
		strings.ToLower(string(chainName)),
		strings.ToLower(string(engineAddr)),
		strings.ToLower(string(pairId)),
		strings.ToLower(string(tradeId)),
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	hashValue := HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
	// fmt.Printf(">> str: %s, hash: %s\n", str, hashValue)
	return hashValue
}

func GetActivityHash(
	chainId DecimalStr,
	chainName string,
	trezorAddr HexStrWith0x,
	activityId DecimalStr,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"trezor-activity|%s|%s|%s|%s",
		strings.ToLower(string(chainId)),
		strings.ToLower(string(chainName)),
		strings.ToLower(string(trezorAddr)),
		strings.ToLower(string(activityId)),
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	hashValue := HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
	// fmt.Printf(">> str: %s, hash: %s\n", str, hashValue)
	return hashValue
}

func GetOrderEventHash(
	orderEventType int,
	orderExchangeHash,
	otherOrderExchangeHash,
	tradeHash HexStrWith0x,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"agora-order-event|%d|%s|%s|%s",
		orderEventType,
		strings.ToLower(string(orderExchangeHash)),
		strings.ToLower(string(otherOrderExchangeHash)),
		strings.ToLower(string(tradeHash)),
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	hashValue := HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
	// fmt.Printf(">> str: %s, hash: %s\n", str, hashValue)
	return hashValue
}

func GetDirectOrderHash(
	chainId DecimalStr,
	chainName string,
	engineAddr HexStrWith0x,
	pairId HexStrWith0x,
	orderId DecimalStr,
	externalOid string,
	exchangeOid string,
) HexStrWith0x {
	hash := sha512.New384()
	str := fmt.Sprintf(
		"agora-direct-order|%s|%s|%s|%s|%s|%s|%s",
		strings.ToLower(string(chainId)),
		strings.ToLower(chainName),
		strings.ToLower(string(engineAddr)),
		strings.ToLower(string(pairId)),
		strings.ToLower(string(orderId)),
		strings.ToLower(externalOid),
		strings.ToLower(exchangeOid),
	)
	hash.Write([]byte(str))
	buf := hash.Sum(nil)
	return HexStrWith0x("0x" + strings.ToLower(hex.EncodeToString(buf)))
}

// HashSHA256 computes SHA-512 hash of the input and truncates to specified number of bytes.
// Returns hex-encoded string with 0x prefix, lowercase (e.g., "0xabc123...")
// The truncation takes the last N bytes of the hash.
// Note: Despite the function name, this uses SHA-512 to support output sizes up to 64 bytes.
func HashSHA256(input string, outputBytes int) string {
	hash := sha512.Sum512([]byte(input))
	// Truncate from the end (last N bytes)
	if outputBytes > len(hash) {
		outputBytes = len(hash)
	}
	truncated := hash[len(hash)-outputBytes:]
	return "0x" + strings.ToLower(hex.EncodeToString(truncated))
}
