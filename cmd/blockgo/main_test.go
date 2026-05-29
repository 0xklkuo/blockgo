package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"blockgo/internal/config"
)

func TestRunGenLocalnetLocalModeGeneratesRunnableConfig(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	if err := runGenLocalnet([]string{"-mode", localnetModeLocal, "-out", outDir}); err != nil {
		t.Fatalf("runGenLocalnet returned error: %v", err)
	}

	cfg, err := config.LoadNodeConfig(filepath.Join(outDir, "node1.json"))
	if err != nil {
		t.Fatalf("LoadNodeConfig returned error: %v", err)
	}

	if got, want := cfg.ListenAddr, "127.0.0.1:7001"; got != want {
		t.Fatalf("ListenAddr = %q, want %q", got, want)
	}

	if got, want := cfg.HTTPAddr, "127.0.0.1:8001"; got != want {
		t.Fatalf("HTTPAddr = %q, want %q", got, want)
	}

	if got, want := cfg.GenesisFile, filepath.Join(outDir, localnetGenesisFileName); got != want {
		t.Fatalf("GenesisFile = %q, want %q", got, want)
	}

	if got, want := cfg.DataDir, "./data/run-node/node1"; got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}

	if _, err := config.LoadGenesis(cfg.GenesisFile); err != nil {
		t.Fatalf("LoadGenesis returned error: %v", err)
	}
}

func TestRunGenLocalnetDockerModePreservesComposeOrientedPaths(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	if err := runGenLocalnet([]string{"-mode", localnetModeDocker, "-out", outDir}); err != nil {
		t.Fatalf("runGenLocalnet returned error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(outDir, "node1.json"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	var cfg localnetNodeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if got, want := cfg.DataDir, "/app/data/node1"; got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}

	if got, want := cfg.GenesisFile, "/app/configs/local/genesis.json"; got != want {
		t.Fatalf("GenesisFile = %q, want %q", got, want)
	}

	if got, want := cfg.ListenAddr, "0.0.0.0:7001"; got != want {
		t.Fatalf("ListenAddr = %q, want %q", got, want)
	}

	if len(cfg.Peers) == 0 || cfg.Peers[0] != "node2:7002" {
		t.Fatalf("Peers = %#v, want first peer node2:7002", cfg.Peers)
	}
}

func TestRunGenLocalnetSingleNodeGeneratesPeerlessConfig(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	if err := runGenLocalnet([]string{"-mode", localnetModeLocal, "-nodes", "1", "-out", outDir}); err != nil {
		t.Fatalf("runGenLocalnet returned error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(outDir, "node1.json"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	var cfg localnetNodeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if len(cfg.Peers) != 0 {
		t.Fatalf("Peers = %#v, want none", cfg.Peers)
	}

	if len(cfg.ValidatorPublicKeys) != 1 {
		t.Fatalf("ValidatorPublicKeys has length %d, want 1", len(cfg.ValidatorPublicKeys))
	}
}

func TestRunGenLocalnetRejectsUnknownMode(t *testing.T) {
	t.Parallel()

	err := runGenLocalnet([]string{"-mode", "nope", "-out", t.TempDir()})
	if err == nil {
		t.Fatal("runGenLocalnet returned nil error, want failure")
	}
}

func TestRunGenLocalnetRejectsInvalidNodeCount(t *testing.T) {
	t.Parallel()

	err := runGenLocalnet([]string{"-nodes", "0", "-out", t.TempDir()})
	if err == nil {
		t.Fatal("runGenLocalnet returned nil error, want failure")
	}
}
