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
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func respondWithError(res http.ResponseWriter, code int, message string) {
	res.WriteHeader(code)
	responseBody := struct {
		Error string `json:"error"`
	}{Error: message}
	dat, err := json.Marshal(responseBody)
	if err != nil {
		log.Println("Error marshaling")
	}
	res.Write(dat)

}

func respondWithJSON(res http.ResponseWriter, code int, payload any) {
	res.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		respondWithError(res, 400, "Cannot marshal json")
	}
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

func validateChirp(chrp database.CreateChirpParams) (database.CreateChirpParams, error) {

	if len(chrp.Body) > 140 {
		return database.CreateChirpParams{}, fmt.Errorf("Chirp to long")
	}
	chrp.Body = profaneFilter(chrp.Body, []string{"kerfuffle", "sharbert", "fornax"})
	return chrp, nil

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
func (cfg *apiConfig) register(res http.ResponseWriter, req *http.Request) {
	requestParams := struct {
		Email string `json:"email"`
	}{}

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(res, 400, "Couldn't decode json")
	}

	usr, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:     requestParams.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        uuid.New()})
	if err != nil {
		fmt.Println(err)
	}
	respondWithJSON(res, 201, usr)

}

func (cfg *apiConfig) reset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.db.Reset(req.Context())
}
func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return
	}
	var srv http.Server
	fileServe := http.StripPrefix("/app", http.FileServer(http.Dir('.')))
	apiCfg := apiConfig{db: database.New(db)}
	ServeMux := http.NewServeMux()
	ServeMux.HandleFunc("POST /api/chirps", apiCfg.setChirp)
	ServeMux.HandleFunc("POST /api/users", apiCfg.register)
	ServeMux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	ServeMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	ServeMux.HandleFunc("GET /api/healthz", Healthy)
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServe))
	srv.Handler = ServeMux
	srv.Addr = ":8080"
	srv.ListenAndServe()
}
