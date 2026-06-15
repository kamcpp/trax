package common

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/quickfixgo/quickfix"
)

const (
	AgoraParticipantOrderIdMagicPrefix = "AGPOID"
)

var encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")

func EncodeToAlphanumeric(input string) string {
	return encoding.EncodeToString([]byte(input))
}

func DecodeFromAlphanumeric(input string) (string, error) {
	decoded, err := encoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func GetParticipantId(sessionID quickfix.SessionID) string {
	// TODO(kam): why target and not sender?
	participantId := sessionID.TargetCompID
	if len(sessionID.TargetSubID) > 0 {
		participantId += ":" + sessionID.TargetSubID
	}
	return participantId
}

func ToParticipantOrderId(
	participantId,
	clientOrderId,
	sideStr,
	symbol string,
) string {
	encodedParticipantId := EncodeToAlphanumeric(participantId)
	encodedClientOrderId := EncodeToAlphanumeric(clientOrderId)
	encodedSymbol := EncodeToAlphanumeric(symbol)
	return EncodeToAlphanumeric(
		fmt.Sprintf("%s|%s|%s|%s|%s",
			AgoraParticipantOrderIdMagicPrefix,
			encodedParticipantId,
			encodedClientOrderId,
			strings.ToUpper(sideStr),
			encodedSymbol,
		),
	)
}

func ParseParticipantOrderId(participantOrderId string) (
	string, // participantId
	string, // clientOrderId
	string, // upper(side)
	string, // symbol
	error,
) {
	decoded, err := base64.StdEncoding.DecodeString(participantOrderId)
	if err != nil {
		return "", "", "", "", err
	}
	tokens := strings.Split(string(decoded), "|")
	if tokens[0] != AgoraParticipantOrderIdMagicPrefix {
		return "", "", "", "",
			fmt.Errorf("invalid agora particiant order id: '%s'", participantOrderId)
	}
	encodedParticipantId := tokens[1]
	encodedClientOrderId := tokens[2]
	upperSideStr := tokens[3]
	encodedSymbol := tokens[4]
	participantId, err := DecodeFromAlphanumeric(encodedParticipantId)
	if err != nil {
		return "", "", "", "", err
	}
	clientOrderId, err := DecodeFromAlphanumeric(encodedClientOrderId)
	if err != nil {
		return "", "", "", "", err
	}
	symbol, err := DecodeFromAlphanumeric(encodedSymbol)
	if err != nil {
		return "", "", "", "", err
	}
	return participantId, clientOrderId, upperSideStr, symbol, nil
}
