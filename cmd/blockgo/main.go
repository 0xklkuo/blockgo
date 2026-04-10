package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
	"blockgo/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "version":
		fmt.Println(version.String())
	case "generate-key", "gen-key":
		if err := runGenerateKey(); err != nil {
			exitErr(err)
		}
	case "address":
		if err := runAddress(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "create-tx":
		if err := runCreateTx(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "gen-localnet":
		if err := runGenLocalnet(os.Args[2:]); err != nil {
			exitErr(err)
		}
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Printf(`%s CLI

Usage:
  blockgo <command>

Commands:
  version
  generate-key
  gen-key
  address -pubkey <hex>
  create-tx -in-prev-tx <hex> -in-index <n> -in-pubkey <hex> -in-privkey <hex> -out-address <hex> -out-value <n> [-change-address <hex> -change-value <n>]
  gen-localnet -out <dir>
`, version.Name)
}

func runGenerateKey() error {
	pub, priv, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		return err
	}

	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		return err
	}

	out := map[string]string{
		"public_key_hex":  hex.EncodeToString(pub),
		"private_key_hex": hex.EncodeToString(priv),
		"address":         addr.String(),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func runAddress(args []string) error {
	fs := flag.NewFlagSet("address", flag.ContinueOnError)
	pubKeyHex := fs.String("pubkey", "", "hex-encoded ed25519 public key")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pubKeyHex == "" {
		return fmt.Errorf("-pubkey is required")
	}

	pub, err := hex.DecodeString(*pubKeyHex)
	if err != nil {
		return err
	}

	addr, err := blockcrypto.AddressFromPublicKey(ed25519.PublicKey(pub))
	if err != nil {
		return err
	}

	fmt.Println(addr.String())
	return nil
}

func runCreateTx(args []string) error {
	fs := flag.NewFlagSet("create-tx", flag.ContinueOnError)

	inPrevTx := fs.String("in-prev-tx", "", "input previous tx id hex")
	inIndex := fs.Uint("in-index", 0, "input output index")
	inPubKey := fs.String("in-pubkey", "", "input public key hex")
	inPrivKey := fs.String("in-privkey", "", "input private key hex")

	outAddress := fs.String("out-address", "", "recipient address hex")
	outValue := fs.Uint64("out-value", 0, "recipient value")

	changeAddress := fs.String("change-address", "", "change address hex")
	changeValue := fs.Uint64("change-value", 0, "change value")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *inPrevTx == "" || *inPubKey == "" || *inPrivKey == "" || *outAddress == "" || *outValue == 0 {
		return fmt.Errorf("required flags are missing")
	}

	prevTxBytes, err := hex.DecodeString(*inPrevTx)
	if err != nil {
		return err
	}
	if len(prevTxBytes) != 32 {
		return fmt.Errorf("invalid prev tx id length")
	}

	pub, err := hex.DecodeString(*inPubKey)
	if err != nil {
		return err
	}

	priv, err := hex.DecodeString(*inPrivKey)
	if err != nil {
		return err
	}

	recipientAddr, err := blockcrypto.ParseAddress(*outAddress)
	if err != nil {
		return err
	}

	var prevHash blockchain.Hash
	copy(prevHash[:], prevTxBytes)

	tx := blockchain.Transaction{
		Inputs: []blockchain.TxInput{
			{
				PrevOut: blockchain.OutPoint{
					TxID:  prevHash,
					Index: uint32(*inIndex),
				},
				PublicKey: pub,
			},
		},
		Outputs: []blockchain.TxOutput{
			{
				Value:   *outValue,
				Address: recipientAddr,
			},
		},
	}

	if *changeAddress != "" && *changeValue > 0 {
		changeAddr, err := blockcrypto.ParseAddress(*changeAddress)
		if err != nil {
			return err
		}

		tx.Outputs = append(tx.Outputs, blockchain.TxOutput{
			Value:   *changeValue,
			Address: changeAddr,
		})
	}

	if err := tx.SignInput(0, ed25519.PrivateKey(priv)); err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(tx)
}

type localnetNodeConfig struct {
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

type localnetGenesis struct {
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

const (
	localnetDefaultOutDir        = "./configs/local"
	localnetNodeCount            = 3
	localnetChainID              = "blockgo-local"
	localnetCreatedAtUTC         = "2026-01-01T00:00:00Z"
	localnetBlockIntervalSeconds = 8
	localnetMaxTxPerBlock        = 1024
	localnetGenesisFileName      = "genesis.json"
	localnetGenesisFilePath      = "/app/configs/local/genesis.json"
	localnetDataDirPattern       = "/app/data/%s"
	localnetListenAddrPattern    = "0.0.0.0:%d"
	localnetNodeIDPattern        = "node%d"
	localnetPeerAddrPattern      = "%s:%d"
	localnetInitialAllocation    = 1000
	localnetP2PPortBase          = 7000
	localnetHTTPPortBase         = 8000
)

func runGenLocalnet(args []string) error {
	fs := flag.NewFlagSet("gen-localnet", flag.ContinueOnError)
	outDir := fs.String("out", localnetDefaultOutDir, "output directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}

	type validatorInfo struct {
		NodeID   string
		Pub      []byte
		Priv     []byte
		Address  string
		P2PPort  int
		HTTPPort int
	}

	validators := make([]validatorInfo, 0, localnetNodeCount)
	for i := 1; i <= localnetNodeCount; i++ {
		pub, priv, err := blockcrypto.GenerateKeyPair()
		if err != nil {
			return err
		}

		addr, err := blockcrypto.AddressFromPublicKey(pub)
		if err != nil {
			return err
		}

		validators = append(validators, validatorInfo{
			NodeID:   fmt.Sprintf(localnetNodeIDPattern, i),
			Pub:      pub,
			Priv:     priv,
			Address:  addr.String(),
			P2PPort:  localnetP2PPortBase + i,
			HTTPPort: localnetHTTPPortBase + i,
		})
	}

	validatorPubKeys := make([]string, 0, len(validators))
	for _, v := range validators {
		validatorPubKeys = append(validatorPubKeys, hex.EncodeToString(v.Pub))
	}

	genesis := localnetGenesis{
		ChainID:              localnetChainID,
		CreatedAtUTC:         localnetCreatedAtUTC,
		BlockIntervalSeconds: localnetBlockIntervalSeconds,
	}
	for _, v := range validators {
		genesis.Validators = append(genesis.Validators, struct {
			ID        string `json:"id"`
			PublicKey string `json:"public_key"`
		}{
			ID:        v.NodeID,
			PublicKey: hex.EncodeToString(v.Pub),
		})
	}
	genesis.Allocations = append(genesis.Allocations, struct {
		Address string `json:"address"`
		Value   uint64 `json:"value"`
	}{
		Address: validators[0].Address,
		Value:   localnetInitialAllocation,
	})

	if err := writeJSONFile(filepath.Join(*outDir, localnetGenesisFileName), genesis); err != nil {
		return err
	}

	for _, v := range validators {
		peers := make([]string, 0, len(validators)-1)
		for _, other := range validators {
			if other.NodeID == v.NodeID {
				continue
			}
			peers = append(peers, fmt.Sprintf(localnetPeerAddrPattern, other.NodeID, other.P2PPort))
		}

		cfg := localnetNodeConfig{
			NodeID:               v.NodeID,
			DataDir:              fmt.Sprintf(localnetDataDirPattern, v.NodeID),
			ListenAddr:           fmt.Sprintf(localnetListenAddrPattern, v.P2PPort),
			HTTPAddr:             fmt.Sprintf(localnetListenAddrPattern, v.HTTPPort),
			Peers:                peers,
			GenesisFile:          localnetGenesisFilePath,
			BlockIntervalSeconds: localnetBlockIntervalSeconds,
			MaxTxPerBlock:        localnetMaxTxPerBlock,
			PrivateKeyHex:        hex.EncodeToString(v.Priv),
			ValidatorPublicKeys:  validatorPubKeys,
		}

		if err := writeJSONFile(filepath.Join(*outDir, v.NodeID+".json"), cfg); err != nil {
			return err
		}
	}

	fmt.Printf("generated localnet config in %s\n", *outDir)
	return nil
}

func writeJSONFile(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o644)
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
