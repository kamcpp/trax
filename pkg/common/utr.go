//go:build ignore
// +build ignore

package common

func GetUTRs() []*ContractDeployment {
	return []*ContractDeployment{ // ethereum mainnet
		{ // polygon mainnet
			ChainId:   137,
			ChainName: "polygon-mainnet",
			Address:   "0xf7d63dba75b1cd117cf4ca84b6540a595e846a77",
		},
	}
}

func GetUTR(chainId uint64) *ContractDeployment {
	for _, utr := range GetUTRs() {
		if utr.ChainId == chainId {
			return utr
		}
	}
	return nil
}
