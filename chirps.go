package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func validateChirp(chrp database.CreateChirpParams) (database.CreateChirpParams, error) {

	if len(chrp.Body) > 140 {
		return database.CreateChirpParams{}, fmt.Errorf("Chirp to long")
	}
	chrp.Body = profaneFilter(chrp.Body, []string{"kerfuffle", "sharbert", "fornax"})
	return chrp, nil

}
func (cfg *apiConfig) getChirp(res http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		fmt.Println(err)
	}
	chrp, err := cfg.db.GetChirp(req.Context(), chirpID)
	if err != nil {
		fmt.Println(err)
		respondWithError(res, http.StatusNotFound, "Chirp down!")
	}
	respondWithJSON(res, http.StatusOK, chrp)
}
func (cfg *apiConfig) getChirps(res http.ResponseWriter, req *http.Request) {
	feed, err := cfg.db.GetChirps(req.Context())
	if err != nil {
		fmt.Println(err)
	}
	respondWithJSON(res, http.StatusOK, feed)
}

func (cfg *apiConfig) setChirp(res http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("%v ", err))
		return
	}

	uid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("%v ", err))
		return
	}

	requestParams := database.CreateChirpParams{UserID: uid, UpdatedAt: time.Now(), CreatedAt: time.Now(), ID: uuid.New()}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("%v ", err))
		return
	}
	chrp, err := validateChirp(requestParams)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("%v ", err))
		return
	}
	chrpDB, err := cfg.db.CreateChirp(req.Context(), chrp)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("%v ", err))
		return
	}
	respondWithJSON(res, http.StatusCreated, chrpDB)
}
