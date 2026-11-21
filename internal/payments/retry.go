package payments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type gatewayRequest struct {
	PaymentID string `json:"payment_id"`
	Amount    int64  `json:"amount_minor"`
	Currency  string `json:"currency"`
}

func (h *handler) charge(r *http.Request, payment paymentRequest, paymentID string) (string, error) {
	body, err := json.Marshal(gatewayRequest{
		PaymentID: paymentID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
	})
	if err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.gatewayURL, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := h.client.Do(request)
		if err != nil {
			lastErr = err
			if attempt < 3 {
				time.Sleep(time.Duration(100*(1<<(attempt-1))) * time.Millisecond)
			}
			continue
		}

		if response.StatusCode >= 200 && response.StatusCode < 300 {
			var result struct {
				GatewayReference string `json:"gateway_reference"`
			}
			err := json.NewDecoder(response.Body).Decode(&result)
			_ = response.Body.Close()
			if err != nil {
				return "", err
			}
			return result.GatewayReference, nil
		}

		status := response.StatusCode
		retryAfter := response.Header.Get("Retry-After")
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		if status >= 400 && status < 500 {
			return "", fmt.Errorf("gateway returned non-retryable status %d", status)
		}

		lastErr = fmt.Errorf("gateway returned status %d", status)
		if attempt < 3 {
			delay := time.Duration(100*(1<<(attempt-1))) * time.Millisecond
			if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
				delay = time.Duration(seconds) * time.Second
				if delay > 2*time.Second {
					delay = 2 * time.Second
				}
			}
			time.Sleep(delay)
		}
	}

	return "", fmt.Errorf("three attempts failed: %w", lastErr)
}