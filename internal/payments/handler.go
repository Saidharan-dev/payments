package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type paymentRequest struct {
	Amount   int64  `json:"amount_minor"`
	Currency string `json:"currency"`
}

type paymentResult struct {
	PaymentID        string `json:"payment_id"`
	Status           string `json:"status"`
	GatewayReference string `json:"gateway_reference"`
}

type handler struct {
	gatewayURL string
	client     *http.Client
	keys       *idempotencyStore
	nextID     atomic.Uint64
}

func NewHandler(gatewayURL string) http.Handler {
	return &handler{
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 3 * time.Second},
		keys:       newIdempotencyStore(),
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		http.Error(w, "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	var request paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))
	if request.Amount <= 0 || (request.Currency != "INR" && request.Currency != "USD") {
		http.Error(w, "amount must be positive and currency must be INR or USD", http.StatusBadRequest)
		return
	}

	result, created, err := h.keys.Do(key, func() (paymentResult, error) {
		paymentID := fmt.Sprintf("pay-%06d", h.nextID.Add(1))
		reference, err := h.charge(r, request, paymentID)
		if err != nil {
			return paymentResult{}, err
		}
		return paymentResult{
			PaymentID:        paymentID,
			Status:           "approved",
			GatewayReference: reference,
		}, nil
	})
	if err != nil {
		http.Error(w, "gateway request failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	}
	_ = json.NewEncoder(w).Encode(result)
}