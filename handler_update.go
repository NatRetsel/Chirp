package main

import (
	"encoding/json"
	"net/http"

	internal "github.com/natretsel/chirpy/internal/auth"
	"github.com/natretsel/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdateInfo(w http.ResponseWriter, r *http.Request) {
	// Get access token from header
	accessToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "malformed header", err)
		return
	}
	// Verify access token, return unauthorized if invalid
	userId, err := internal.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid access token", err)
		return
	}

	// Unmarshal request body
	type userParameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	reqBody := userParameters{}
	err = decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to decode request body", err)
		return
	}
	// Get user by ID
	userDBObj, err := cfg.dbQueries.GetUserByID(r.Context(), userId)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user", err)
	}
	// hash password, check if it's the same pw
	hashedPW, err := internal.HashPassword(reqBody.Password)
	if err != nil || hashedPW == userDBObj.HashedPassword {
		respondWithError(w, http.StatusBadRequest, "Please use a different password", err)
	}

	// update in DB and respond with updated user resource
	updatedUser, err := cfg.dbQueries.UpdateLoginDetailsByID(r.Context(), database.UpdateLoginDetailsByIDParams{
		HashedPassword: hashedPW,
		Email:          reqBody.Email,
		ID:             userId,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update password in DB", err)
		return
	}

	type response struct {
		User
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:            updatedUser.ID,
			CreatedAt:     updatedUser.CreatedAt,
			UpdatedAt:     updatedUser.UpdatedAt,
			Email:         updatedUser.Email,
			Is_chirpy_red: updatedUser.IsChirpyRed.Bool,
		},
	})
}
