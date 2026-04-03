package crypto

import "testing"

func TestGenerateKeyPairAndSignVerify(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	msg := []byte("blockgo-test-message")

	sig, err := Sign(priv, msg)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !Verify(pub, msg, sig) {
		t.Fatal("Verify() = false, want true")
	}
}

func TestVerifyRejectsWrongMessage(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	sig, err := Sign(priv, []byte("message-a"))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if Verify(pub, []byte("message-b"), sig) {
		t.Fatal("Verify() = true, want false")
	}
}
