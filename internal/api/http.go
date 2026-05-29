package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"blockgo/internal/blockchain"
	blockcrypto "blockgo/internal/crypto"
)

const (
	contentTypeJSON            = "application/json"
	healthStatusOK             = "ok"
	routeHealth                = "GET /healthz"
	routeChainHead             = "GET /v1/chain/head"
	routeMempool               = "GET /v1/mempool"
	routeSubmitTransaction     = "POST /v1/transactions"
	maxTransactionRequestBytes = 1 << 20
)

type NodeAPI interface {
	Head() *blockchain.Block
	MempoolLen() int
	SubmitTransaction(tx blockchain.Transaction) error
}

type Server struct {
	logger *slog.Logger
	node   NodeAPI
	mux    *http.ServeMux
}

type healthResponse struct {
	Status string `json:"status"`
}

type mempoolResponse struct {
	Size int `json:"size"`
}

type submitTransactionResponse struct {
	TxID string `json:"tx_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewServer(logger *slog.Logger, node NodeAPI) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		logger: logger,
		node:   node,
		mux:    http.NewServeMux(),
	}

	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc(routeHealth, s.handleHealth)
	s.mux.HandleFunc(routeChainHead, s.handleHead)
	s.mux.HandleFunc(routeMempool, s.handleMempool)
	s.mux.HandleFunc(routeSubmitTransaction, s.handleSubmitTransaction)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: healthStatusOK,
	})
}

func (s *Server) handleHead(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.node.Head())
}

func (s *Server) handleMempool(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, mempoolResponse{
		Size: s.node.MempoolLen(),
	})
}

type submitTransactionRequest struct {
	Inputs []struct {
		PrevTxID     string `json:"prev_tx_id"`
		OutputIndex  uint32 `json:"output_index"`
		PublicKeyHex string `json:"public_key_hex"`
		SignatureHex string `json:"signature_hex"`
	} `json:"inputs"`
	Outputs []struct {
		Value   uint64 `json:"value"`
		Address string `json:"address"`
	} `json:"outputs"`
}

func (s *Server) handleSubmitTransaction(w http.ResponseWriter, r *http.Request) {
	req, err := decodeSubmitTransactionRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tx, err := decodeTransactionRequest(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := s.node.SubmitTransaction(tx); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusAccepted, submitTransactionResponse{
		TxID: tx.ID.String(),
	})
}

func decodeSubmitTransactionRequest(w http.ResponseWriter, r *http.Request) (submitTransactionRequest, error) {
	var req submitTransactionRequest

	r.Body = http.MaxBytesReader(w, r.Body, maxTransactionRequestBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		return req, err
	}

	var extra struct{}
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return req, errors.New("request body must contain a single JSON object")
		}
		return req, err
	}

	return req, nil
}

func decodeTransactionRequest(req submitTransactionRequest) (blockchain.Transaction, error) {
	tx := blockchain.Transaction{
		Inputs:  make([]blockchain.TxInput, 0, len(req.Inputs)),
		Outputs: make([]blockchain.TxOutput, 0, len(req.Outputs)),
	}

	for _, in := range req.Inputs {
		prevTxIDBytes, err := hex.DecodeString(in.PrevTxID)
		if err != nil {
			return tx, err
		}
		if len(prevTxIDBytes) != 32 {
			return tx, errors.New("invalid prev_tx_id length")
		}

		pub, err := hex.DecodeString(in.PublicKeyHex)
		if err != nil {
			return tx, err
		}

		sig, err := hex.DecodeString(in.SignatureHex)
		if err != nil {
			return tx, err
		}

		var prevHash blockchain.Hash
		copy(prevHash[:], prevTxIDBytes)

		tx.Inputs = append(tx.Inputs, blockchain.TxInput{
			PrevOut: blockchain.OutPoint{
				TxID:  prevHash,
				Index: in.OutputIndex,
			},
			PublicKey: pub,
			Signature: sig,
		})
	}

	for _, out := range req.Outputs {
		addr, err := blockcrypto.ParseAddress(out.Address)
		if err != nil {
			return tx, err
		}

		tx.Outputs = append(tx.Outputs, blockchain.TxOutput{
			Value:   out.Value,
			Address: addr,
		})
	}

	if err := tx.Finalize(); err != nil {
		return tx, err
	}

	return tx, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{
		Error: err.Error(),
	})
}
