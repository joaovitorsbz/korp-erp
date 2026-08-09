package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total de requisicoes HTTP recebidas pelo http-server-projeto-korp",
	},
	[]string{"path", "method", "status"},
)

type projetoKorpResponse struct {
	Nome    string `json:"nome"`
	Horario string `json:"horario"`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func withMetrics(path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)
		httpRequestsTotal.WithLabelValues(path, r.Method, strconv.Itoa(rec.status)).Inc()
	}
}

func projetoKorpHandler(w http.ResponseWriter, r *http.Request) {
	resp := projetoKorpResponse{
		Nome:    "Projeto Korp",
		Horario: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/projeto-korp", withMetrics("/projeto-korp", projetoKorpHandler))
	mux.Handle("/metrics", promhttp.Handler())

	const addr = ":8080"
	log.Printf("http-server-projeto-korp escutando em %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
