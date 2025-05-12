package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type chirp struct {
	Body string `json:"body"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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
		responseBody := struct {
			Error string `json:"error"`
		}{Error: "Something went wrong"}
		dat, err := json.Marshal(responseBody)
		if err != nil {
			log.Println("Error marshaling")
		}
		res.WriteHeader(http.StatusBadRequest)
		res.Write(dat)
	} else if len(params.Body) > 140 {
		responseBody := struct {
			Error string `json:"error"`
		}{Error: "Chirp is to long"}
		res.WriteHeader(http.StatusBadRequest)
		dat, err := json.Marshal(responseBody)
		if err != nil {
			log.Println("Error marshaling")
		}
		res.Write(dat)
	} else {
		responseBody := struct {
			Valid bool `json:"valid"`
		}{Valid: true}
		res.WriteHeader(http.StatusOK)
		dat, err := json.Marshal(responseBody)
		if err != nil {
			log.Println("Error marshaling")
		}
		res.Write(dat)
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

func (cfg *apiConfig) reset(res http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
}
func main() {
	var srv http.Server
	fileServe := http.StripPrefix("/app", http.FileServer(http.Dir('.')))
	apiCfg := apiConfig{}
	ServeMux := http.NewServeMux()
	ServeMux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	ServeMux.HandleFunc("POST /admin/reset", apiCfg.reset)
	ServeMux.HandleFunc("GET /api/healthz", Healthy)
	ServeMux.HandleFunc("POST /api/validate_chirp", validateChirp)
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServe))
	srv.Handler = ServeMux
	srv.Addr = ":8080"
	srv.ListenAndServe()
}
