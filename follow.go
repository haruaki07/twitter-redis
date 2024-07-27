package main

import (
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// /follow?uid=<user_id>&f=1
func follow(w http.ResponseWriter, r *http.Request) {
	authUser := getRequestUser(r.Context())
	if authUser == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	uid := r.URL.Query().Get("uid")
	if uid == "" {
		http.Error(w, "parameter uid is required", http.StatusBadRequest)
		return
	}

	f := r.URL.Query().Get("f")
	if f == "" {
		http.Error(w, "parameter f is required", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	if uid != authUser.Id {
		if f == "1" {
			err := rc.ZAdd(r.Context(), "followers:"+uid, redis.Z{Score: float64(now), Member: authUser.Id}).Err()
			if err != nil {
				log.Printf("failed to add user's follower: %e\n", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			err = rc.ZAdd(r.Context(), "following:"+authUser.Id, redis.Z{Score: float64(now), Member: uid}).Err()
			if err != nil {
				log.Printf("failed to add user to following list: %e\n", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		} else if f == "0" {
			err := rc.ZRem(r.Context(), "followers:"+uid, authUser.Id).Err()
			if err != nil {
				log.Printf("failed to remove user's follower: %e\n", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			err = rc.ZRem(r.Context(), "following:"+authUser.Id, uid).Err()
			if err != nil {
				log.Printf("failed to remove user from following list: %e\n", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
