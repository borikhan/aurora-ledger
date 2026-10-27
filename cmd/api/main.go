package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/borikhan/aurora-ledger/internal/ledger"
)

func main() {
	port := env("PORT", "8080")
	book := ledger.New()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("POST /v1/accounts", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ID, Currency        string
			OpeningBalanceCents int64
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			write(w, 400, map[string]string{"error": "invalid json"})
			return
		}
		a, err := book.CreateAccount(in.ID, in.Currency, in.OpeningBalanceCents)
		if err != nil {
			write(w, 422, map[string]string{"error": err.Error()})
			return
		}
		write(w, 201, a)
	})
	mux.HandleFunc("POST /v1/transfers", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			ID, From, To string
			AmountCents  int64
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			write(w, 400, map[string]string{"error": "invalid json"})
			return
		}
		t, err := book.Transfer(in.ID, in.From, in.To, in.AmountCents)
		if err != nil {
			write(w, 422, map[string]string{"error": err.Error()})
			return
		}
		write(w, 201, t)
	})
	mux.HandleFunc("GET /v1/accounts/{id}", func(w http.ResponseWriter, r *http.Request) {
		a, err := book.Account(r.PathValue("id"))
		if err != nil {
			write(w, 404, map[string]string{"error": err.Error()})
			return
		}
		write(w, 200, a)
	})
	srv := &http.Server{Addr: ":" + port, Handler: requestLog(mux), ReadHeaderTimeout: 5 * time.Second}
	slog.Info("ledger api listening", "port", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
