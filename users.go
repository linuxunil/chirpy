package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Password  string `json:"password"`
	Email     string `json:"email"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

func (cfg *apiConfig) register(res http.ResponseWriter, req *http.Request) {
	var requestParams User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(res, 400, fmt.Sprintf("Couldn't decode json: %v", err))
	}

	hashed_pass, err := auth.HashPassword(requestParams.Password)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, fmt.Sprintf("Password hashing: %v", err))
	}
	usr, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          requestParams.Email,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		HashedPassword: hashed_pass,
		ID:             uuid.New()})
	if err != nil {
		respondWithError(res, 400, fmt.Sprintf("Error creating user: %v", err))
	}
	respondWithJSON(res, 201, usr)

}

func (cfg *apiConfig) login(res http.ResponseWriter, req *http.Request) {
	var requestParams User
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&requestParams)
	if err != nil {
		respondWithError(res, 400, fmt.Sprintf("Couldn't decode json: %v", err))
	}
	// Get users ID
	usr, err := cfg.db.GetUserAndPassByName(req.Context(), requestParams.Email)
	if err != nil {
		respondWithError(res, 404, fmt.Sprintf("User not found: %v", err))
	}
	// Check passwords
	err = auth.CheckPasswordHash(requestParams.Password, usr.HashedPassword)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Invalid password: %v", err))
	}
	// retUsr, err := cfg.db.GetUserByName(req.Context(), requestParams.Email)
	// if err != nil {
	// 	respondWithError(res, 404, fmt.Sprintf("User not found: %v", err))
	// }
	expiry := 1 * time.Hour
	if requestParams.ExpiresIn != 0 {
		expiry = time.Duration(requestParams.ExpiresIn * int(time.Second))
	}
	token, err := auth.MakeJWT(usr.ID, cfg.secret, expiry)
	refresh, err := auth.MakeRefreshToken()
	body := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
		Refresh   string    `json:"refresh"`
	}{ID: usr.ID, CreatedAt: time.Now(), UpdatedAt: time.Now(), Email: usr.Email, Token: token, Refresh: refresh}
	_, err = cfg.db.CreateRefresh(req.Context(),
		database.CreateRefreshParams{Token: token,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    usr.ID,
			ExpiresAt: time.Now().Add(time.Hour * 1440),
		})
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Failed to write refresh token: %v", err))
	}
	respondWithJSON(res, 200, body)
}

func (cfg *apiConfig) refresh(res http.ResponseWriter, req *http.Request) {
	token, _ := auth.GetBearerToken(req.Header)
	refresh, err := cfg.db.GetToken(req.Context(), token)
	if err != nil {
		respondWithError(res, 401, "Unauthorized")
	}
	if refresh.ExpiresAt.Before(time.Now()) {
		cfg.db.RevokeToken(context.Background(), database.RevokeTokenParams{Token: token, RevokedAt: sql.NullTime{Time: time.Now(), Valid: true}})
		respondWithError(res, 401, "Unauthorized")
		return
	}
	respondWithJSON(res, 200, struct {
		Token string `json:"token"`
	}{Token: token})

}
func (cfg *apiConfig) revoke(res http.ResponseWriter, req *http.Request) {
	token, _ := auth.GetBearerToken(req.Header)
	cfg.db.RevokeToken(context.Background(), database.RevokeTokenParams{Token: token, RevokedAt: sql.NullTime{Time: time.Now(), Valid: true}})
	res.WriteHeader(204)
	res.Write(nil)

}
