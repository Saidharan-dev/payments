package main

import (
	"log"
	"net/http"
	"os"

	paymenthttp "github.com/Saidharan-dev/payments/internal/payments"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8082/charge"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("payments ok"))
	})
	mux.Handle("/payments", paymenthttp.NewHandler(gatewayURL))

	log.Printf("payments listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}