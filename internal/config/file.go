package config

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

type nodeConfigFile struct {
	NodeID               string   `json:"node_id"`
	DataDir              string   `json:"data_dir"`
	ListenAddr           string   `json:"listen_addr"`
	HTTPAddr             string   `json:"http_addr"`
	Peers                []string `json:"peers"`
	GenesisFile          string   `json:"genesis_file"`
	BlockIntervalSeconds int      `json:"block_interval_seconds"`
	MaxTxPerBlock        int      `json:"max_tx_per_block"`
	PrivateKeyHex        string   `json:"private_key_hex"`
	ValidatorPublicKeys  []string `json:"validator_public_keys"`
}

type genesisFile struct {
	ChainID              string `json:"chain_id"`
	CreatedAtUTC         string `json:"created_at_utc"`
	BlockIntervalSeconds int    `json:"block_interval_seconds"`
	Validators           []struct {
		ID        string `json:"id"`
		PublicKey string `json:"public_key"`
	} `json:"validators"`
	Allocations []struct {
		Address string `json:"address"`
		Value   uint64 `json:"value"`
	} `json:"allocations"`
}

type LoadedNodeConfig struct {
	NodeConfig
	HTTPAddr    string
	GenesisFile string
}

func LoadNodeConfig(path string) (LoadedNodeConfig, error) {
	var out LoadedNodeConfig

	raw, err := os.ReadFile(path)
	if err != nil {
		return out, fmt.Errorf("read node config: %w", err)
	}

	var file nodeConfigFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return out, fmt.Errorf("unmarshal node config: %w", err)
	}

	priv, err := hex.DecodeString(file.PrivateKeyHex)
	if err != nil {
		return out, fmt.Errorf("decode private key hex: %w", err)
	}

	validators := make([][]byte, 0, len(file.ValidatorPublicKeys))
	for i, keyHex := range file.ValidatorPublicKeys {
		key, err := hex.DecodeString(keyHex)
		if err != nil {
			return out, fmt.Errorf("decode validator public key %d: %w", i, err)
		}
		validators = append(validators, key)
	}

	out = LoadedNodeConfig{
		NodeConfig: NodeConfig{
			NodeID:              file.NodeID,
			DataDir:             file.DataDir,
			ListenAddr:          file.ListenAddr,
			Peers:               file.Peers,
			BlockInterval:       time.Duration(file.BlockIntervalSeconds) * time.Second,
			MaxTxPerBlock:       file.MaxTxPerBlock,
			LocalPrivateKey:     priv,
			ValidatorPublicKeys: validators,
		},
		HTTPAddr:    file.HTTPAddr,
		GenesisFile: resolveRelativePath(path, file.GenesisFile),
	}

	if err := out.NodeConfig.NormalizeAndValidate(); err != nil {
		return out, err
	}

	if out.GenesisFile == "" {
		return out, fmt.Errorf("genesis_file is required")
	}

	return out, nil
}

func LoadGenesis(path string) (*blockchain.Block, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read genesis file: %w", err)
	}

	var file genesisFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("unmarshal genesis file: %w", err)
	}

	validators := make([][]byte, 0, len(file.Validators))
	for i, v := range file.Validators {
		pub, err := hex.DecodeString(v.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("decode validator public key %d: %w", i, err)
		}
		validators = append(validators, pub)
	}

	allocations := make([]blockchain.GenesisAllocation, 0, len(file.Allocations))
	for i, a := range file.Allocations {
		addr, err := blockcrypto.ParseAddress(a.Address)
		if err != nil {
			return nil, fmt.Errorf("parse allocation address %d: %w", i, err)
		}

		allocations = append(allocations, blockchain.GenesisAllocation{
			Address: addr,
			Value:   a.Value,
		})
	}

	genesis, err := blockchain.NewGenesisBlock(blockchain.GenesisConfig{
		ChainID:     file.ChainID,
		Validators:  validators,
		Allocations: allocations,
	})
	if err != nil {
		return nil, fmt.Errorf("build genesis block: %w", err)
	}

	return genesis, nil
}

func resolveRelativePath(baseFile, target string) string {
	if target == "" {
		return ""
	}
	if filepath.IsAbs(target) {
		return target
	}
	return filepath.Join(filepath.Dir(baseFile), target)
}
