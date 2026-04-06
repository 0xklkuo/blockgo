package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"blockgo/internal/blockchain"

	bolt "go.etcd.io/bbolt"
	bboltErrors "go.etcd.io/bbolt/errors"
)

var (
	blocksBucketName = []byte("blocks")
	utxosBucketName  = []byte("utxos")
	metaBucketName   = []byte("meta")

	headHashKey   = []byte("head_hash")
	headHeightKey = []byte("head_height")
)

type Store struct {
	db *bolt.DB
}

type persistedUTXO struct {
	TxID   blockchain.Hash     `json:"tx_id"`
	Index  uint32              `json:"index"`
	Output blockchain.TxOutput `json:"output"`
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("db path is required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open bbolt db: %w", err)
	}

	s := &Store{db: db}

	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range [][]byte{blocksBucketName, utxosBucketName, metaBucketName} {
			if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
				return fmt.Errorf("create bucket %q: %w", string(bucketName), err)
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.Close()
}

func (s *Store) SaveGenesis(block *blockchain.Block, utxos *blockchain.UTXOSet) error {
	if block == nil {
		return errors.New("nil genesis block")
	}

	if block.Header.Height != 0 {
		return errors.New("genesis block height must be 0")
	}

	if utxos == nil {
		return errors.New("nil utxo set")
	}

	return s.saveBlockAndState(block, utxos)
}

func (s *Store) SaveBlock(block *blockchain.Block, utxos *blockchain.UTXOSet) error {
	if block == nil {
		return errors.New("nil block")
	}

	if utxos == nil {
		return errors.New("nil utxo set")
	}

	return s.saveBlockAndState(block, utxos)
}

func (s *Store) saveBlockAndState(block *blockchain.Block, utxos *blockchain.UTXOSet) error {
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("marshal block: %w", err)
	}

	utxoEntries := utxos.Entries()

	return s.db.Update(func(tx *bolt.Tx) error {
		blocksBucket := tx.Bucket(blocksBucketName)
		utxosBucket := tx.Bucket(utxosBucketName)
		metaBucket := tx.Bucket(metaBucketName)

		if err := blocksBucket.Put(block.Hash[:], blockBytes); err != nil {
			return fmt.Errorf("store block: %w", err)
		}

		if err := clearBucket(utxosBucket); err != nil {
			return fmt.Errorf("clear utxos bucket: %w", err)
		}

		for key, output := range utxoEntries {
			record := persistedUTXO{
				TxID:   key.TxID,
				Index:  key.Index,
				Output: output,
			}

			value, err := json.Marshal(record)
			if err != nil {
				return fmt.Errorf("marshal utxo: %w", err)
			}

			if err := utxosBucket.Put(encodeUTXOKey(key), value); err != nil {
				return fmt.Errorf("store utxo: %w", err)
			}
		}

		if err := metaBucket.Put(headHashKey, block.Hash[:]); err != nil {
			return fmt.Errorf("store head hash: %w", err)
		}

		heightBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(heightBytes, block.Header.Height)

		if err := metaBucket.Put(headHeightKey, heightBytes); err != nil {
			return fmt.Errorf("store head height: %w", err)
		}

		return nil
	})
}

func (s *Store) LoadHead() (*blockchain.Block, error) {
	var headHash blockchain.Hash
	found := false

	err := s.db.View(func(tx *bolt.Tx) error {
		metaBucket := tx.Bucket(metaBucketName)
		raw := metaBucket.Get(headHashKey)
		if len(raw) == 0 {
			return nil
		}

		if len(raw) != len(headHash) {
			return errors.New("invalid stored head hash length")
		}

		copy(headHash[:], raw)
		found = true
		return nil
	})
	if err != nil {
		if errors.Is(err, bboltErrors.ErrBucketNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("load head hash: %w", err)
	}

	if !found {
		return nil, nil
	}

	return s.LoadBlock(headHash)
}

func (s *Store) LoadBlock(hash blockchain.Hash) (*blockchain.Block, error) {
	var block *blockchain.Block

	err := s.db.View(func(tx *bolt.Tx) error {
		blocksBucket := tx.Bucket(blocksBucketName)
		raw := blocksBucket.Get(hash[:])
		if len(raw) == 0 {
			return nil
		}

		var decoded blockchain.Block
		if err := json.Unmarshal(raw, &decoded); err != nil {
			return fmt.Errorf("unmarshal block: %w", err)
		}

		block = &decoded
		return nil
	})
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (s *Store) LoadUTXOSet() (*blockchain.UTXOSet, error) {
	utxos := blockchain.NewUTXOSet()

	err := s.db.View(func(tx *bolt.Tx) error {
		utxosBucket := tx.Bucket(utxosBucketName)

		return utxosBucket.ForEach(func(_, value []byte) error {
			var record persistedUTXO
			if err := json.Unmarshal(value, &record); err != nil {
				return fmt.Errorf("unmarshal utxo: %w", err)
			}

			utxos.Put(blockchain.OutPoint{
				TxID:  record.TxID,
				Index: record.Index,
			}, record.Output)

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

func clearBucket(bucket *bolt.Bucket) error {
	cursor := bucket.Cursor()

	for key, _ := cursor.First(); key != nil; key, _ = cursor.Next() {
		if err := cursor.Delete(); err != nil {
			return err
		}
	}

	return nil
}

func encodeUTXOKey(key blockchain.UTXOKey) []byte {
	out := make([]byte, 36)
	copy(out[:32], key.TxID[:])
	binary.BigEndian.PutUint32(out[32:], key.Index)
	return out
}
