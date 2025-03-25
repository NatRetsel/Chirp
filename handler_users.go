package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// Unmarshal the json, retrieve email from body
	type userParameters struct {
		Email string `json:"email"`
	}
	type userResponse struct {
		User
	}
	decoder := json.NewDecoder(r.Body)
	userParam := userParameters{}
	err := decoder.Decode(&userParam)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	// check for existing user

	/*
		user, err := cfg.dbQueries.GetUserByEmail(r.Context(), userParam.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't query user", err)
		}

		if user.Email == userParam.Email {
			respondWithError(w, http.StatusBadRequest, "User with email already exist", nil)
		}
	*/

	// create user in DB with the email
	user, err := cfg.dbQueries.CreateUser(r.Context(), userParam.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't create user", err)
		return
	}

	// if successfully created, api response with code 201
	userJSON := userResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	}

	respondWithJSON(w, http.StatusCreated, userJSON)
}
