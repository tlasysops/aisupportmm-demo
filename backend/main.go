package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// GitCommit is injected at build time: -ldflags "-X main.GitCommit=abc123"
var GitCommit = "dev"

type server struct {
	db *sql.DB
}

type item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// /api/health - no DB needed (container liveness)
func (s *server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// /api/info - shows version + DB connectivity
func (s *server) info(w http.ResponseWriter, r *http.Request) {
	host, _ := os.Hostname()
	dbStatus := "connected"
	if s.db == nil {
		dbStatus = "not configured"
	} else if err := s.db.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}
	writeJSON(w, 200, map[string]string{
		"commit": GitCommit, "hostname": host, "db": dbStatus,
	})
}

func (s *server) items(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		writeJSON(w, 503, map[string]string{"error": "database not configured"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		rows, err := s.db.Query(`SELECT id, name, created_at FROM items ORDER BY id DESC LIMIT 50`)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()
		out := []item{}
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.ID, &it.Name, &it.CreatedAt); err != nil {
				writeJSON(w, 500, map[string]string{"error": err.Error()})
				return
			}
			out = append(out, it)
		}
		writeJSON(w, 200, out)
	case http.MethodPost:
		var in struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Name == "" {
			writeJSON(w, 400, map[string]string{"error": "name is required"})
			return
		}
		var it item
		err := s.db.QueryRow(
			`INSERT INTO items (name) VALUES ($1) RETURNING id, name, created_at`, in.Name,
		).Scan(&it.ID, &it.Name, &it.CreatedAt)
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 201, it)
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

// cors lets the frontend domain call this API from the browser.
func cors(next http.Handler) http.Handler {
	origin := os.Getenv("ALLOWED_ORIGIN")
	if origin == "" {
		origin = "https://weblayer.aisupportmm.com"
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Vary", "Origin")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.health)
	mux.HandleFunc("/api/info", s.info)
	mux.HandleFunc("/api/items", s.items)
	return cors(mux)
}

func connectDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	for i := 1; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("waiting for db (%d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Printf("db unavailable, starting anyway: %v", err)
		return db
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now())`)
	if err != nil {
		log.Printf("migrate: %v", err)
	}
	return db
}

func main() {
	s := &server{}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		s.db = connectDB(dsn)
	} else {
		log.Println("DATABASE_URL not set; running without database")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("backend %s listening on :%s", GitCommit, port)
	log.Fatal(http.ListenAndServe(":"+port, s.routes()))
}
