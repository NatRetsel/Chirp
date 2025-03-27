package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	internal "github.com/natretsel/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpgradeChirpyRed(w http.ResponseWriter, r *http.Request) {

	// check API key
	reqAPIKey, err := internal.GetAPIKey(r.Header)
	if reqAPIKey != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "invalid API Key", err)
	}
	type RequestData struct {
		UserID uuid.UUID `json:"user_id"`
	}
	type RequestParam struct {
		Event string      `json:"event"`
		Data  RequestData `json:"data"`
	}
	// unmarshall request Param
	decoder := json.NewDecoder(r.Body)
	requestParam := RequestParam{}
	err = decoder.Decode(&requestParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't unmarshall request", err)
		return
	}

	// check if event is "user.upgraded"; return 204 if it is not
	if requestParam.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// update user in database, mark they are chirpy red member
	_, err = cfg.dbQueries.UpgradeChirpyRedByID(r.Context(), requestParam.Data.UserID)
	// return 404 if user not found (err not nil)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "invalid user", err)
		return
	}

	// return 204
	w.WriteHeader(http.StatusNoContent)
}
