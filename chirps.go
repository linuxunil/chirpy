package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) webHooks(res http.ResponseWriter, req *http.Request) {
	auth, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(res, 401, "Invalid api")
		return
	}
	if auth != os.Getenv("POLKA_KEY") {
		respondWithError(res, 401, "Wrong Key")
		return
	}
	request := struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}{}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&request)
	if err != nil {
		respondWithError(res, 401, "Invalid json")
		return
	}
	if request.Event != "user.upgraded" {
		respondWithJSON(res, 204, "")
		return
	}
	err = cfg.db.MarkRed(req.Context(), request.Data.UserID)
	fmt.Println("Upgraded to red", request.Data.UserID)
	if err != nil {
		respondWithJSON(res, 404, "")
	}
	respondWithJSON(res, 204, "")
}
func (cfg *apiConfig) rmChirp(res http.ResponseWriter, req *http.Request) {
	chirpID, _ := uuid.Parse(req.PathValue("chirpID"))
	bearerToken, _ := auth.GetBearerToken(req.Header)
	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(res, 401, "invalid token")
		return
	}
	chrp, err := cfg.db.GetChirp(req.Context(), chirpID)
	if err != nil {
		respondWithError(res, 404, "Chirp not found")
		return
	}
	if chrp.UserID != userID {
		respondWithError(res, 403, "Not Authorized")
		return
	}
	err = cfg.db.RmChirp(context.Background(), chrp.ID)
	if err != nil {
		respondWithError(res, 401, "Unable to delete")
	}
	respondWithJSON(res, 204, "")

}
func validateChirp(chrp database.CreateChirpParams) (database.CreateChirpParams, error) {

	if len(chrp.Body) > 140 {
		return database.CreateChirpParams{}, fmt.Errorf("Chirp to long")
	}
	chrp.Body = profaneFilter(chrp.Body, []string{"kerfuffle", "sharbert", "fornax"})
	return chrp, nil

}
func (cfg *apiConfig) getChirp(res http.ResponseWriter, req *http.Request) {
	usrID, err := uuid.Parse(req.PathValue("userID"))
	feed, err := cfg.db.GetChiprsByUserID(req.Context(), usrID)
	if err != nil {
		fmt.Println(err)
	}
	respondWithJSON(res, http.StatusOK, feed)

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
		respondWithError(res, 401, fmt.Sprintf("Get Bearer %v ", err))
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
