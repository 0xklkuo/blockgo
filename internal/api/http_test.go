package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

type stubNode struct {
	head      *blockchain.Block
	mempool   int
	submitErr error
	lastTx    blockchain.Transaction
}

func (n *stubNode) Head() *blockchain.Block {
	return n.head
}

func (n *stubNode) MempoolLen() int {
	return n.mempool
}

func (n *stubNode) SubmitTransaction(tx blockchain.Transaction) error {
	n.lastTx = tx
	return n.submitErr
}

func TestSubmitTransactionAcceptsValidRequest(t *testing.T) {
	t.Parallel()

	node := &stubNode{}
	server := NewServer(nil, node)

	body := bytes.NewBufferString(validTransactionRequestJSON(t))
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", body)
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	if got, want := resp.Code, http.StatusAccepted; got != want {
		t.Fatalf("status = %d, want %d, body = %s", got, want, resp.Body.String())
	}

	var out submitTransactionResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if out.TxID == "" {
		t.Fatal("expected tx_id in response")
	}
	if blockchain.IsZeroHash(node.lastTx.ID) {
		t.Fatal("expected submitted transaction to be finalized")
	}
}

func TestSubmitTransactionRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	server := NewServer(nil, &stubNode{})
	body := strings.NewReader(`{"inputs":[],"outputs":[],"unexpected":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", body)
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	assertErrorResponseContains(t, resp, http.StatusBadRequest, "unknown field")
}

func TestSubmitTransactionRejectsTrailingJSON(t *testing.T) {
	t.Parallel()

	server := NewServer(nil, &stubNode{})
	body := strings.NewReader(validTransactionRequestJSON(t) + `{}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", body)
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	assertErrorResponseContains(t, resp, http.StatusBadRequest, "single JSON object")
}

func TestSubmitTransactionRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	server := NewServer(nil, &stubNode{})
	oversizedAddress := strings.Repeat("a", maxTransactionRequestBytes)
	body := strings.NewReader(`{"inputs":[],"outputs":[{"value":1,"address":"` + oversizedAddress + `"}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", body)
	resp := httptest.NewRecorder()

	server.Handler().ServeHTTP(resp, req)

	assertErrorResponseContains(t, resp, http.StatusBadRequest, "request body too large")
}

func assertErrorResponseContains(t *testing.T, resp *httptest.ResponseRecorder, wantStatus int, wantSubstring string) {
	t.Helper()

	if got := resp.Code; got != wantStatus {
		t.Fatalf("status = %d, want %d, body = %s", got, wantStatus, resp.Body.String())
	}

	var out errorResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if !strings.Contains(out.Error, wantSubstring) {
		t.Fatalf("error = %q, want substring %q", out.Error, wantSubstring)
	}
}

func validTransactionRequestJSON(t *testing.T) string {
	t.Helper()

	pub, _, err := blockcrypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}
	addr, err := blockcrypto.AddressFromPublicKey(pub)
	if err != nil {
		t.Fatalf("AddressFromPublicKey returned error: %v", err)
	}

	payload := map[string]any{
		"inputs": []map[string]any{
			{
				"prev_tx_id":     strings.Repeat("0", 64),
				"output_index":   0,
				"public_key_hex": hex.EncodeToString(pub),
				"signature_hex":  hex.EncodeToString(make([]byte, ed25519.SignatureSize)),
			},
		},
		"outputs": []map[string]any{
			{
				"value":   1,
				"address": addr.String(),
			},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	return string(raw)
}
