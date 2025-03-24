package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type cleanedTextResponse struct {
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	const maxChirpLength = 140

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	bannedWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	cleanedBody := getCleanedBody(params.Body, bannedWords)

	respondWithJSON(w, 200, cleanedTextResponse{
		Body: cleanedBody,
	})

}

func getCleanedBody(str string, bannedWords map[string]bool) string {
	redactedString := "****"
	splitBody := strings.Split(str, " ")
	for idx, word := range splitBody {
		if _, ok := bannedWords[strings.ToLower(word)]; ok {
			splitBody[idx] = redactedString

		}
	}
	return strings.Join(splitBody, " ")
}
