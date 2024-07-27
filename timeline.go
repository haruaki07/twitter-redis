package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func timeline(w http.ResponseWriter, r *http.Request) {
	limit := 50

	// retrieve the timeline, which contains posts id, start and stop is the element index
	ids, err := rc.LRange(r.Context(), "timeline", 0, int64(limit-1)).Result()
	if err != nil {
		log.Printf("failed to retrieve timeline ids: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// map timeline posts detail
	posts := make([]*Post, 0, limit)
	for _, id := range ids {
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

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
