package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rushilp164/go_rss_demo/internal/database"
)

func (cfg *apiConfig) handlerGetPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10

	if specifiedLimit, err := strconv.Atoi(limitStr); err == nil {
		limit = specifiedLimit
	}

	posts, err := cfg.DB.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})

	if err != nil {
		respondWithError(w, http.StatusOK, fmt.Sprintf("Couldn't fetch feeds: %v", err))
	}

	respondWithJSON(w, http.StatusOK, posts)
}
