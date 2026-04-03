package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

const AddressLength = 32

type Address [AddressLength]byte

func AddressFromPublicKey(publicKey ed25519.PublicKey) (Address, error) {
	var addr Address

	if len(publicKey) != ed25519.PublicKeySize {
		return addr, errors.New("invalid ed25519 public key length")
	}

	sum := sha256.Sum256(publicKey)
	copy(addr[:], sum[:])

	return addr, nil
}

func MustAddressFromPublicKey(publicKey ed25519.PublicKey) Address {
	addr, err := AddressFromPublicKey(publicKey)
	if err != nil {
		panic(err)
	}

	return addr
}

func ParseAddress(s string) (Address, error) {
	var addr Address

	raw, err := hex.DecodeString(s)
	if err != nil {
		return addr, err
	}

	if len(raw) != AddressLength {
		return addr, errors.New("invalid address length")
	}

	copy(addr[:], raw)
	return addr, nil
}

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}

func (a Address) Bytes() []byte {
	out := make([]byte, len(a))
	copy(out, a[:])
	return out
}

func (a Address) IsZero() bool {
	var zero Address
	return a == zero
}
