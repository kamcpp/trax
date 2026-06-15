//go:build ignore
// +build ignore

package common

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
)

type ContractDeployment struct {
	ChainId   uint64 `json:"chain_id"`
	ChainName string `json:"chain_name"`
	Address   string `json:"address"`
}

type Client struct {
	ApiKey      string                `json:"api_key"`
	Orderbooks  []*ContractDeployment `json:"orderbooks"`
	Trezors     []*ContractDeployment `json:"trezors"`
	WebhookURLs []string              `json:"webhook_urls"`
}

func GetClients() map[string]*Client {
	apiKey := os.Getenv("AGORA_API_KEY")
	if len(apiKey) == 0 {
		panic("API_KEY is not set")
	}
	chainIdStr := os.Getenv("CHAIN_ID")
	if len(chainIdStr) == 0 {
		panic("CHAIN_ID is not set")
	}
	chainId := MustStrToUint64(DecimalStr(chainIdStr))
	chainName := os.Getenv("CHAIN_NAME")
	if len(chainName) == 0 {
		panic("CHAIN_NAME is not set")
	}
	engineAddr := os.Getenv("ENGINE_ADDR")
	if len(engineAddr) == 0 {
		panic("ENGINE_ADDR is not set")
	}
	trezorAddr := os.Getenv("TREZOR_ADDR")
	if len(trezorAddr) == 0 {
		panic("TREZOR_ADDR is not set")
	}
	clientsMap := make(map[string]*Client)
	{
		clientsMap["public"] = &Client{ // TODO(kam): remove this later when F2 is in place
			ApiKey: "public",
			Orderbooks: []*ContractDeployment{
				{
					ChainId:   chainId,
					ChainName: chainName,
					Address:   engineAddr,
				},
			},
			Trezors: []*ContractDeployment{
				{
					ChainId:   chainId,
					ChainName: chainName,
					Address:   trezorAddr,
				},
			},
			WebhookURLs: []string{
				"https://",
			},
		}
	}
	{
		clientsMap[apiKey] = &Client{
			ApiKey: apiKey,
			Orderbooks: []*ContractDeployment{
				{
					ChainId:   chainId,
					ChainName: chainName,
					Address:   engineAddr,
				},
			},
			Trezors: []*ContractDeployment{
				{
					ChainId:   chainId,
					ChainName: chainName,
					Address:   trezorAddr,
				},
			},
			WebhookURLs: []string{
				"https://",
			},
		}
	}
	return clientsMap
}

func GetClientFromGinContext(c *gin.Context) (*Client, error) {
	var apiKey string
	apiKey = c.GetHeader("x-agora-api-key")
	if len(apiKey) == 0 {
		var found bool
		apiKey, found = c.GetQuery("agora-api-key")
		if !found {
			return nil, errors.New("api-key header not found")
		}
	}
	clients := GetClients()
	client, ok := clients[apiKey]
	if !ok {
		L.Warn("no client found using the given api-key", F(c)...)
		c.JSON(401, gin.H{})
		c.Abort()
		return nil, errors.New("no client found using the given api-key")
	}
	return client, nil
}
