package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ProfileOut struct {
	UserId      string  `json:"user_id"`
	Username    string  `json:"username"`
	IsFollowing *bool   `json:"is_following,omitempty"`
	Posts       []*Post `json:"posts"`
}

func profile(w http.ResponseWriter, r *http.Request) {
	authUser := getRequestUser(r.Context())
	if authUser == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username param is required", http.StatusBadRequest)
		return
	}

	// get requested user id by username
	// data is an hash which contain username as a fields
	// users { username = id, username = id, ... }
	userId, err := rc.HGet(r.Context(), "users", username).Result()
	if err != nil {
		log.Printf("failed to get user id: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// get user's list of product ids
	postIds, err := rc.LRange(r.Context(), "posts:"+userId, 0, 9).Result()
	if err != nil {
		log.Printf("failed to get post ids: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// map posts
	posts := make([]*Post, 0, 10)
	for _, id := range postIds {
		post := &Post{}
		// get the post detail
		fields, err := rc.HGetAll(r.Context(), "post:"+id).Result()
		if err != nil {
			log.Printf("failed to get post detail: %e\n", err)
			continue
		}

		post.Id = id
		post.Body = fields["body"]
		post.Time = fields["time"]
		post.UserId = fields["user_id"]

		// get the post author's username
		username, err := rc.HGet(r.Context(), "user:"+post.UserId, "username").Result()
		if err != nil {
			log.Printf("failed to get username: %e\n", err)
		}

		post.Username = username

		posts = append(posts, post)
	}

	out := ProfileOut{
		UserId:   userId,
		Username: username,
		Posts:    posts,
	}

	if authUser.Id != userId {
		// get the following status from sorted set by member value
		// ZScore will return error if there are no such member
		// following:user_id [{ score = time, member = user_id }, ...]
		_, err = rc.ZScore(r.Context(), "following:"+userId, authUser.Id).Result()
		isFollowing := err == nil
		*out.IsFollowing = isFollowing
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(out)
}
