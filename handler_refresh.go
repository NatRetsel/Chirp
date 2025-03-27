package main

import (
	"net/http"
	"time"

	internal "github.com/natretsel/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerValidateRefresh(w http.ResponseWriter, r *http.Request) {
	// check for Bearer token in Header
	refreshToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No bearer token", err)
		return
	}
	// Look up refresh token in DB
	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
		return
	}

	// otherwise return 200 and {"token":"{access token}"}
	jwtToken, err := internal.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	type AccessTokenResp struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, AccessTokenResp{
		Token: jwtToken,
	})
}
