package config

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"time"
)

const (
	DefaultBlockInterval = 5 * time.Second
	DefaultMaxTxPerBlock = 1000
)

type NodeConfig struct {
	NodeID              string
	DataDir             string
	ListenAddr          string
	Peers               []string
	BlockInterval       time.Duration
	MaxTxPerBlock       int
	LocalPrivateKey     ed25519.PrivateKey
	ValidatorPublicKeys [][]byte
}

func (c *NodeConfig) NormalizeAndValidate() error {
	if c == nil {
		return errors.New("nil node config")
	}

	if c.NodeID == "" {
		return errors.New("node id is required")
	}

	if c.DataDir == "" {
		return errors.New("data dir is required")
	}

	if c.ListenAddr == "" {
		return errors.New("listen address is required")
	}

	if len(c.LocalPrivateKey) != ed25519.PrivateKeySize {
		return errors.New("invalid local private key length")
	}

	if len(c.ValidatorPublicKeys) == 0 {
		return errors.New("at least one validator public key is required")
	}

	for i, key := range c.ValidatorPublicKeys {
		if len(key) != ed25519.PublicKeySize {
			return fmt.Errorf("validator public key at index %d has invalid length", i)
		}
	}

	if c.BlockInterval <= 0 {
		c.BlockInterval = DefaultBlockInterval
	}

	if c.MaxTxPerBlock <= 0 {
		c.MaxTxPerBlock = DefaultMaxTxPerBlock
	}

	return nil
}

func (c NodeConfig) LocalPublicKey() ed25519.PublicKey {
	pub := c.LocalPrivateKey.Public().(ed25519.PublicKey)
	out := make([]byte, len(pub))
	copy(out, pub)
	return out
}
