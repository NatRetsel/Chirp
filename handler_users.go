package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	internal "github.com/natretsel/chirpy/internal/auth"
	"github.com/natretsel/chirpy/internal/database"
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
		Email    string `json:"email"`
		Password string `json:"password"`
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
	hashedPassword, err := internal.HashPassword(userParam.Password)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "please use a different password", err)
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userParam.Email,
		HashedPassword: hashedPassword,
	})
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
