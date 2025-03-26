package main

import (
	"encoding/json"
	"net/http"
	"time"

	internal "github.com/natretsel/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type userParameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	decoder := json.NewDecoder(r.Body)
	loginParam := userParameters{}
	err := decoder.Decode(&loginParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error decoding json", err)
		return
	}

	// Query for user with email
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), loginParam.Email)

	// if user doesn't exist or errors, return 401
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Incorrect email or password", err)
		return
	}
	// compare password hash, return 401 if fails, 200 with copy of user resource otherwise
	err = internal.CheckPasswordHash(user.HashedPassword, loginParam.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	timeToExpiry := time.Hour
	if loginParam.ExpiresInSeconds > 0 && loginParam.ExpiresInSeconds < 3600 {
		timeToExpiry = time.Duration(loginParam.ExpiresInSeconds) * time.Second
	}

	jwtToken, err := internal.MakeJWT(user.ID, cfg.secret, timeToExpiry)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating jwtToken", err)
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token: jwtToken,
	})
}
