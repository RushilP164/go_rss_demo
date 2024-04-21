package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rushilp164/go_rss_demo/internal/database"
)

// Que: What if we don't use pointer here?
func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Name string
	}
	decoder := json.NewDecoder(r.Body)
	params := userParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Couldn't decode parameters: %v", err))
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		log.Fatal(err)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}
