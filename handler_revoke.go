package main

import (
	"net/http"

	internal "github.com/natretsel/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Malformed header", err)
		return
	}
	err = cfg.dbQueries.RevokeToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error revoking refresh token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
