package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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
