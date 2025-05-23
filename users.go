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

func (cfg *apiConfig) updateUser(res http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Bearer Token: %v", err))
		return
	}

	usr, err := auth.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Invalid access Token: %v", err))
		return
	}

	body := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&body)
	if err != nil {
		respondWithError(res, 401, "no email pass in body")
		return
	}
	hashed, _ := auth.HashPassword(body.Password)
	upUser, err := cfg.db.UpdateUser(context.Background(), database.UpdateUserParams{Email: body.Email, HashedPassword: hashed, ID: usr})

	respondWithJSON(res, 200, upUser)

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

	// Set refresh token
	expiry := 1 * time.Hour // Default to hour expiration
	// If expiratoin is set overwrite
	if requestParams.ExpiresIn != 0 {
		expiry = time.Duration(requestParams.ExpiresIn * int(time.Second))
	}

	token, err := auth.MakeJWT(usr.ID, cfg.secret, expiry)

	refresh, err := auth.MakeRefreshToken()

	rt, err := cfg.db.CreateRefresh(req.Context(),

		database.CreateRefreshParams{Token: refresh,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    usr.ID,
			ExpiresAt: time.Now().Add(time.Hour * 1440),
		})
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Failed to write refresh token: %v\n\t%v\n",
			err, rt))
	}

	user, err := cfg.db.GetUserByID(req.Context(), usr.ID)
	body := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		Token       string    `json:"token"`
		Refresh     string    `json:"refresh_token"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{ID: usr.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Email:       user.Email,
		Token:       token,
		Refresh:     rt.Token,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(res, 200, body)
}

func (cfg *apiConfig) refresh(res http.ResponseWriter, req *http.Request) {
	headerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Println(err)
		respondWithError(res, 401, fmt.Sprintf("Get bearer in refresh %v", err))
		return
	}

	refreshToken, err := cfg.db.GetToken(req.Context(), headerToken)
	fmt.Println("Token :", headerToken)
	if err != nil {
		respondWithError(res, 401, fmt.Sprintf("Refresh not found %v", err))
		return
	}

	if refreshToken.RevokedAt.Valid {
		respondWithError(res, 401, fmt.Sprintf("Revoked refresh: %v", refreshToken.RevokedAt))
	}
	expired := refreshToken.ExpiresAt.Local().Before(time.Now())

	if expired {
		cfg.db.RevokeToken(context.Background(), database.RevokeTokenParams{
			Token: headerToken,
			RevokedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true}})

		respondWithError(res, 401, fmt.Sprintf("Token expired %v", refreshToken.ExpiresAt))
		return
	}

	tokenAuth, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(res, 401, "Generate JWT")
	}
	respondWithJSON(res, 200, struct {
		Token string `json:"token"`
	}{Token: tokenAuth})

}
func (cfg *apiConfig) revoke(res http.ResponseWriter, req *http.Request) {
	token, _ := auth.GetBearerToken(req.Header)
	cfg.db.RevokeToken(context.Background(), database.RevokeTokenParams{Token: token, RevokedAt: sql.NullTime{Time: time.Now(), Valid: true}})
	res.WriteHeader(204)
	res.Write(nil)

}
