package main

import (
	"net/http"

	"github.com/google/uuid"
	internal "github.com/natretsel/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	// Get access token, check access token
	accessToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "malformed header", err)
		return
	}

	// check access token
	userId, err := internal.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}
	// Retrieve chirp by ID
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID", err)
		return
	}
	chirpDBObj, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get chirp by ID", err)
		return
	}
	// Check if Chirp is by user through user_id
	if chirpDBObj.UserID != userId {
		respondWithError(w, http.StatusForbidden, "not owner of chirp", err)
		return
	}
	// Delete chirp
	err = cfg.dbQueries.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "couldn't delete chirp by ID", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
