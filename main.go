package main

import (
	"fmt"
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
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServe))
	srv.Handler = ServeMux
	srv.Addr = ":8080"
	srv.ListenAndServe()
}
