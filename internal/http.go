package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func StartHTTPServer(pool *pgxpool.Pool, port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ok"
		dbStatus := "ok"
		if err := pool.Ping(ctx); err != nil {
			status = "fail"
			dbStatus = "fail"
		}

		w.Header().Set("Content-Type", "application/json")
		code := http.StatusOK
		if status != "ok" {
			code = http.StatusServiceUnavailable
		}
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{
			"status":   status,
			"dbStatus": dbStatus,
		})
	})

	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limit := 20
		offset := 0
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if n, err := strconv.Atoi(o); err == nil && n >= 0 {
				offset = n
			}
		}
		nameFilter := r.URL.Query().Get("name")

		tags, err := GetTags(r.Context(), pool, limit, offset, nameFilter)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tags); err != nil {
			http.Error(w, "Encoding error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/api/whatsnews", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		limit := 20
		offset := 0
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		if o := r.URL.Query().Get("offset"); o != "" {
			if n, err := strconv.Atoi(o); err == nil && n >= 0 {
				offset = n
			}
		}

		tagIDs := []int{}
		if ids := r.URL.Query().Get("tags"); ids != "" {
			for _, p := range strings.Split(ids, ",") {
				if id, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
					tagIDs = append(tagIDs, id)
				}
			}
		}

		result, err := GetWhatsnews(r.Context(), pool, limit, offset, tagIDs)
		if err != nil {
			http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
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
