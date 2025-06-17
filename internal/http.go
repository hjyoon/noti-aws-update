package internal

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func StartHTTPServer(pool *pgxpool.Pool, port string) {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexPath := filepath.Join("static", "index.html")
		data, err := os.ReadFile(indexPath)
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
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

		search := r.URL.Query().Get("search")

		result, err := GetWhatsnews(r.Context(), pool, limit, offset, tagIDs, search)
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

func getRealIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	fwd := r.Header.Get("Forwarded")
	if fwd != "" {
		for _, field := range strings.Split(fwd, ";") {
			field = strings.TrimSpace(field)
			if strings.HasPrefix(field, "for=") {
				ip := strings.TrimPrefix(field, "for=")
				ip = strings.Trim(ip, "\"")
				if ip != "" {
					return ip
				}
			}
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}
	return r.RemoteAddr
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
		ip := getRealIP(r)
		log.Printf(
			"[%s] \"%s %s%s\" %d \"%s\" (duration: %s)",
			ip,
			r.Method,
			r.URL.Path,
			query,
			crw.statusCode,
			r.UserAgent(),
			duration,
		)

	})
}
