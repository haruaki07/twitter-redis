package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RegisterIn struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(w http.ResponseWriter, r *http.Request) {
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

	if rc.HGet(r.Context(), "users", req.Username).Val() != "" {
		http.Error(w, "username already exists", http.StatusBadRequest)
		return
	}

	userId, err := rc.Incr(r.Context(), "next_user_id").Result()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	authSecret := randString(32)
	err = rc.HSet(r.Context(), "users", req.Username, userId).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = rc.HSet(r.Context(), fmt.Sprintf("user:%d", userId), map[string]interface{}{
		"username": req.Username,
		"password": req.Password,
		"auth":     authSecret,
	}).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = rc.HSet(r.Context(), "auths", authSecret, userId).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = rc.ZAdd(r.Context(), "users_by_time", redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: req.Username,
	}).Err()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "auth",
		Value:   authSecret,
		Expires: time.Now().Add(time.Hour * 24 * 365), // a year
	})
	w.WriteHeader(http.StatusCreated)
}
