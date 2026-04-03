package crypto

import "testing"

func TestAddressFromPublicKey(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	if addr.IsZero() {
		t.Fatal("address is zero")
	}
}

func TestParseAddressRoundTrip(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	addr, err := AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey() error = %v", err)
	}

	parsed, err := ParseAddress(addr.String())
	if err != nil {
		t.Fatalf("ParseAddress() error = %v", err)
	}

	if parsed != addr {
		t.Fatalf("parsed address = %v, want %v", parsed, addr)
	}
}
