package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/AlRowne/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	Platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
	</html>`
	fmt.Fprintf(w, body, hits)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "forbidden", nil)
		return
	}
	err := cfg.dbQueries.DeleteUsers(r.Context())
	fmt.Println("All users have been deleted")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't delete users", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func main() {
	const port = "8080"
	const filepathRoot = "."

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	var apiCfg apiConfig
	apiCfg.dbQueries = dbQueries
	apiCfg.Platform = os.Getenv("PLATFORM")

	handlerRoot := http.StripPrefix("/app/", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handlerRoot))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUsers)

	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("starting server on http://localhost%v\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(http.StatusText(http.StatusOK)))
}
