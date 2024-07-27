package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type PostIn struct {
	Status string `json:"status"`
}

func post(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req PostIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	status := strings.ReplaceAll(req.Status, "\n", " ")

	postId, err := rc.Incr(r.Context(), "next_post_id").Result()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = rc.HMSet(
		r.Context(),
		fmt.Sprintf("post:%d", postId),
		"user_id", user.Id,
		"time", time.Now().Unix(),
		"body", status,
	).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	followers, err := rc.ZRange(r.Context(), "followers:"+user.Id, 0, -1).Result()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// insert the post in the author's posts list as well
	followers = append(followers, user.Id)

	for _, fid := range followers {
		if err := rc.LPush(r.Context(), "posts:"+fid, postId).Err(); err != nil {
			log.Printf("failed to add post: %e\n", err)
		}
	}

	if err = rc.LPush(r.Context(), "timeline", postId).Err(); err != nil {
		log.Printf("failed to add timeline: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err = rc.LTrim(r.Context(), "timeline", 0, 1000).Err(); err != nil {
		log.Printf("failed to trim timeline: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
