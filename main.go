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

type chirp struct {
	Body string `json:"body"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func respondWithError(res http.ResponseWriter, code int, message string) {
	res.WriteHeader(code)
	responseBody := struct {
		Error string `json:"error"`
	}{Error: "Something went wrong"}
	dat, err := json.Marshal(responseBody)
	if err != nil {
		log.Println("Error marshaling")
	}
	res.Write(dat)

}

func respondWithJSON(res http.ResponseWriter, code int, payload interface{}) {
	res.WriteHeader(code)
	if p, ok := payload.([]byte); ok {
		res.Write(p)
	}

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
func validateChirp(res http.ResponseWriter, req *http.Request) {
	type requestParams struct {
		Body string `json:"body"`
	}

	res.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(req.Body)
	params := requestParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, 400, "Something went wrong")
	} else if len(params.Body) > 140 {
		respondWithError(res, 400, "Chirp is to long")
	} else {
		filtered := profaneFilter(params.Body, []string{"kerfuffle", "sharbert", "fornax"})
		responseBody := struct {
			Cleaned_body string `json:"cleaned_body"`
		}{Cleaned_body: filtered}
		dat, err := json.Marshal(responseBody)
		if err != nil {
			log.Println("Error marshaling")
		}
		respondWithJSON(res, 200, dat)
	}

}
func Healthy(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}
func (cfg *apiConfig) metrics(res http.ResponseWriter, req *http.Request) {
	template := []byte(fmt.Sprintf(`<html>
	  <body>
	    <h1>Welcome, Chirpy Admin</h1>
	    <p>Chirpy has been visited %d times!</p>
	  </body>
	</html>`, cfg.fileserverHits.Load()))
	res.Header().Set("Content-Type", "text/html; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	// body := []byte("Hits: ")
	// body = fmt.Appendf(body, "%v", cfg.fileserverHits.Load())
	res.Write(template)
}
func (cfg *apiConfig) register(res http.ResponseWriter, req *http.Request) {
	requestParams := struct {
		Email string `json:"email"`
	}{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestParams)
	if err != nil {
		log.Fatal(err)
	}

	usr, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:     sql.NullString{String: requestParams.Email, Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ID:        uuid.New()})

	res.WriteHeader(201)
	res.Header().Set("Content-Type", "application/json")
	user := struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email     string `json:"email"`
	}{ID: usr.ID.String(),
		CreatedAt: usr.CreatedAt.String(),
		UpdatedAt: usr.UpdatedAt.GoString(),
		Email:     usr.Email.String}
	dat, err := json.Marshal(user)
	res.Write(dat)

}

func (cfg *apiConfig) reset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
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
	ServeMux.HandleFunc("POST /api/users", apiConfig.register)
	ServeMux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	ServeMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	ServeMux.HandleFunc("GET /api/healthz", Healthy)
	ServeMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServe))
	srv.Handler = ServeMux
	srv.Addr = ":8080"
	srv.ListenAndServe()
}
