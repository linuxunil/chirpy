package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	secret         string
}

func respondWithError(res http.ResponseWriter, code int, message string) {
	responseBody := struct {
		Error string `json:"error"`
	}{Error: message}
	dat, err := json.Marshal(responseBody)
	if err != nil {
		log.Println("Error marshaling")
	}
	res.WriteHeader(code)
	res.Write(dat)

}

func respondWithJSON(res http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(res, 400, "Cannot marshal json")
	}
	res.WriteHeader(code)
	res.Write(dat)

}
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func profaneFilter(unclean string, sins []string) string {
	clean := strings.Split(unclean, " ")

	for i := range clean {
		for j := range sins {
			if strings.ToLower(clean[i]) == sins[j] {
				clean[i] = "****"
			}
		}
	}
	return strings.Join(clean, " ")
}

func Healthy(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}
func (cfg *apiConfig) metrics(res http.ResponseWriter, req *http.Request) {
	var template []byte
	template = fmt.Appendf(template, `<html>
	  <body>
	    <h1>Welcome, Chirpy Admin</h1>
	    <p>Chirpy has been visited %d times!</p>
	  </body>
	</html>`, cfg.fileserverHits.Load())
	res.Header().Set("Content-Type", "text/html; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	res.Write(template)
}

func (cfg *apiConfig) reset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.db.Reset(req.Context())
}
func main() {
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECERET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return
	}
	var srv http.Server
	fileServe := http.StripPrefix("/app", http.FileServer(http.Dir('.')))
	apiCfg := apiConfig{db: database.New(db), secret: secret}
	ServeMux := http.NewServeMux()
	ServeMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	ServeMux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	ServeMux.HandleFunc("POST /api/chirps", apiCfg.setChirp)
	ServeMux.HandleFunc("POST /api/revoke", apiCfg.revoke)
	ServeMux.HandleFunc("POST /api/refresh", apiCfg.refresh)
	ServeMux.HandleFunc("POST /api/login", apiCfg.login)
	ServeMux.HandleFunc("POST /api/users", apiCfg.register)
	ServeMux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	ServeMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	ServeMux.HandleFunc("GET /api/healthz", Healthy)
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServe))
	srv.Handler = ServeMux
	srv.Addr = ":8080"
	srv.ListenAndServe()
}
