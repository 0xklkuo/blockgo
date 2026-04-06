package consensus

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

type ValidatorSet struct {
	validators []ed25519.PublicKey
}

func NewValidatorSet(publicKeys [][]byte) (*ValidatorSet, error) {
	if len(publicKeys) == 0 {
		return nil, errors.New("validator set cannot be empty")
	}

	validators := make([]ed25519.PublicKey, 0, len(publicKeys))
	for i, k := range publicKeys {
		if len(k) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid validator key length at index %d", i)
		}
		key := make([]byte, len(k))
		copy(key, k)
		validators = append(validators, key)
	}

	return &ValidatorSet{validators: validators}, nil
}

func (s *ValidatorSet) Size() int {
	if s == nil {
		return 0
	}
	return len(s.validators)
}

func (s *ValidatorSet) IsValidator(pub []byte) bool {
	if s == nil || len(pub) != ed25519.PublicKeySize {
		return false
	}

	for _, v := range s.validators {
		if bytes.Equal(v, pub) {
			return true
		}
	}

	return false
}

func (s *ValidatorSet) Proposer(height uint64) (ed25519.PublicKey, error) {
	if s == nil || len(s.validators) == 0 {
		return nil, errors.New("empty validator set")
	}

	// height 1 => index 0 (first validator proposes first non-genesis block)
	index := int((height - 1) % uint64(len(s.validators)))
	proposer := make([]byte, len(s.validators[index]))
	copy(proposer, s.validators[index])
	return proposer, nil
}

func (s *ValidatorSet) SignBlock(block *blockchain.Block, privateKey ed25519.PrivateKey) error {
	if s == nil {
		return errors.New("nil validator set")
	}
	if block == nil {
		return errors.New("nil block")
	}
	if len(privateKey) != ed25519.PrivateKeySize {
		return errors.New("invalid private key length")
	}
	if blockchain.IsZeroHash(block.Hash) {
		return errors.New("block hash is zero; finalize block first")
	}

	sig, err := blockcrypto.Sign(privateKey, block.Hash[:])
	if err != nil {
		return err
	}

	block.Signature = sig
	return nil
}

func (s *ValidatorSet) ValidateBlock(block *blockchain.Block) error {
	if s == nil {
		return errors.New("nil validator set")
	}
	if block == nil {
		return errors.New("nil block")
	}
	if len(block.Header.Validator) != ed25519.PublicKeySize {
		return errors.New("invalid block validator public key length")
	}
	if len(block.Signature) != ed25519.SignatureSize {
		return errors.New("invalid block signature length")
	}
	if blockchain.IsZeroHash(block.Hash) {
		return errors.New("block hash is zero")
	}

	expectedProposer, err := s.Proposer(block.Header.Height)
	if err != nil {
		return err
	}

	if !bytes.Equal(block.Header.Validator, expectedProposer) {
		return errors.New("block proposer does not match expected validator rotation")
	}

	if !ed25519.Verify(ed25519.PublicKey(block.Header.Validator), block.Hash[:], block.Signature) {
		return errors.New("invalid block signature")
	}

	return nil
}
