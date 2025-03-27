package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	internal "github.com/natretsel/chirpy/internal/auth"
	"github.com/natretsel/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsGetByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID", err)
	}
	chirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", nil)
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirpID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	// check for optional query param
	authorIDStr := r.URL.Query().Get("author_id")
	order := r.URL.Query().Get("sort")
	// retrieves all chirps in ascending order
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't query all chirps", err)
		return
	}
	authorID := uuid.Nil
	if len(authorIDStr) != 0 {
		authorID, err = uuid.Parse(authorIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author_id", err)
			return
		}
		user, err := cfg.dbQueries.GetUserByID(r.Context(), authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author", err)
			return
		}
		chirps, err = cfg.dbQueries.GetChirpsByUserID(r.Context(), user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't query all chirps", err)
			return
		}
	}

	chirpsArr := []Chirp{}
	for _, c := range chirps {
		if authorID != uuid.Nil && authorID != c.UserID {
			continue
		}
		chirpsArr = append(chirpsArr, Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		})
	}
	if order == "desc" {
		slices.Reverse(chirpsArr)
	}
	respondWithJSON(w, http.StatusOK, chirpsArr)
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type cleanedTextResponse struct {
		Chirp
	}

	bearerToken, err := internal.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "error getting bearer token", err)
		return
	}

	userId, err := internal.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirpParam := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userId,
	}
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), chirpParam)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
	}
	respondWithJSON(w, 201, cleanedTextResponse{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})

}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
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
