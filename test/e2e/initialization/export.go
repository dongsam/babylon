package initialization

import (
	"fmt"
)

type ChainMeta struct {
	DataDir string `json:"dataDir"`
	Id      string `json:"id"`
}

type Node struct {
	Name      string `json:"name"`
	ConfigDir string `json:"configDir"`
	Mnemonic  string `json:"mnemonic"`
	// PublicAddress is validator bech32 AccAddress.String()
	PublicAddress string `json:"publicAddress"`
	WalletName    string `json:"walletName"`
	PublicKey     []byte `json:"publicKey"`
	PrivateKey    []byte `json:"privateKey"`
	PeerId        string `json:"peerId"`
	IsValidator   bool   `json:"isValidator"`
	CometPrivKey  []byte `json:"cometPrivKey"`
}

type Chain struct {
	ChainMeta ChainMeta `json:"chainMeta"`
	Nodes     []*Node   `json:"validators"`
}

func (c *ChainMeta) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.Id)
}
