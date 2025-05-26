package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func StartHTTPServer(conn *pgx.Conn, port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := conn.Ping(ctx); err != nil {
			http.Error(w, "DB NOT OK: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":" + port
	log.Printf("Start Server: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, LoggingMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// CustomResponseWriter는 응답 상태 코드를 기록
type CustomResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *CustomResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		crw := &CustomResponseWriter{w, http.StatusOK}

		next.ServeHTTP(crw, r)

		duration := time.Since(start)
		query := ""
		if r.URL.RawQuery != "" {
			query = "?" + r.URL.RawQuery
		}
		log.Printf(
			"[%s] \"%s %s%s\" %d \"%s\" (duration: %s)",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			query,
			crw.statusCode,
			r.UserAgent(),
			duration,
		)

	})
}
