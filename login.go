package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type LoginIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	var req RegisterIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	// this will get the user id by username
	// the data is an hash that use the username as a key and user id as a value
	// hget will return nil if element not found
	userId, err := rc.HGet(r.Context(), "users", req.Username).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			http.Error(w, "wrong username or password", http.StatusUnauthorized)
			return
		}

		log.Printf("failed to get user id by username: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// this will get the user's password and auth secret
	fields, err := rc.HMGet(r.Context(), "user:"+userId, "password", "auth").Result()
	if err != nil {
		log.Printf("failed to get user password: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if req.Password != fields[0] {
		http.Error(w, "wrong username or password", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "auth",
		Value:   fields[1].(string),
		Expires: time.Now().Add(time.Hour * 24 * 365), // a year
	})
	w.WriteHeader(http.StatusOK)
}
