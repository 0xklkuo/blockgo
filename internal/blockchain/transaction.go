package blockchain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	blockcrypto "blockgo/internal/crypto"
)

type Hash [32]byte

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Bytes() []byte {
	out := make([]byte, len(h))
	copy(out, h[:])
	return out
}

func IsZeroHash(h Hash) bool {
	var zero Hash
	return h == zero
}

type OutPoint struct {
	TxID  Hash   `json:"tx_id"`
	Index uint32 `json:"index"`
}

type TxInput struct {
	PrevOut   OutPoint `json:"prev_out"`
	PublicKey []byte   `json:"public_key"`
	Signature []byte   `json:"signature"`
}

type TxOutput struct {
	Value   uint64              `json:"value"`
	Address blockcrypto.Address `json:"address"`
}

type Transaction struct {
	ID      Hash       `json:"id"`
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

type txHashPayloadInput struct {
	PrevOut   OutPoint `json:"prev_out"`
	PublicKey []byte   `json:"public_key"`
}

type txHashPayload struct {
	Inputs  []txHashPayloadInput `json:"inputs"`
	Outputs []TxOutput           `json:"outputs"`
}

func (tx Transaction) HashPayload() ([]byte, error) {
	inputs := make([]txHashPayloadInput, 0, len(tx.Inputs))
	for _, in := range tx.Inputs {
		inputs = append(inputs, txHashPayloadInput{
			PrevOut:   in.PrevOut,
			PublicKey: in.PublicKey,
		})
	}

	payload := txHashPayload{
		Inputs:  inputs,
		Outputs: tx.Outputs,
	}

	return json.Marshal(payload)
}

func (tx Transaction) ComputeID() (Hash, error) {
	var zero Hash

	payload, err := tx.HashPayload()
	if err != nil {
		return zero, fmt.Errorf("marshal tx hash payload: %w", err)
	}

	sum := sha256.Sum256(payload)
	return Hash(sum), nil
}

func (tx *Transaction) Finalize() error {
	if tx == nil {
		return errors.New("nil transaction")
	}

	id, err := tx.ComputeID()
	if err != nil {
		return err
	}

	tx.ID = id
	return nil
}

func (tx Transaction) SigningMessage() ([]byte, error) {
	return tx.HashPayload()
}

func (tx *Transaction) SignInput(inputIndex int, privateKey ed25519.PrivateKey) error {
	if tx == nil {
		return errors.New("nil transaction")
	}

	if inputIndex < 0 || inputIndex >= len(tx.Inputs) {
		return errors.New("input index out of range")
	}

	msg, err := tx.SigningMessage()
	if err != nil {
		return err
	}

	sig, err := blockcrypto.Sign(privateKey, msg)
	if err != nil {
		return err
	}

	tx.Inputs[inputIndex].Signature = sig
	return nil
}

func (tx Transaction) VerifyInputSignature(inputIndex int) bool {
	if inputIndex < 0 || inputIndex >= len(tx.Inputs) {
		return false
	}

	msg, err := tx.SigningMessage()
	if err != nil {
		return false
	}

	in := tx.Inputs[inputIndex]
	return blockcrypto.Verify(ed25519.PublicKey(in.PublicKey), msg, in.Signature)
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 0 && len(tx.Outputs) > 0
}
