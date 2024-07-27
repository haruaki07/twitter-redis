package main

import (
	"log"
	"net/http"
)

func logout(w http.ResponseWriter, r *http.Request) {
	user := getRequestUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// get the original (old) secret
	oldSecret, err := rc.HGet(r.Context(), "user:"+user.Id, "auth").Result()
	if err != nil {
		log.Printf("failed to get the old secret: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// we need to update the auth secret after logout
	authSecret := randString(32)

	// update the new auth secret to the user hash
	if rc.HSet(r.Context(), "user:"+user.Id, "auth", authSecret).Err() != nil {
		log.Printf("failed to update user auth secret: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// set the new auth secret hash to the auths hash
	if rc.HSet(r.Context(), "auths", authSecret, user.Id).Err() != nil {
		log.Printf("failed to set the new auth secret: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// remove the old auth secret from auths hash
	if rc.HDel(r.Context(), "auths", oldSecret).Err() != nil {
		log.Printf("failed to remove old auth secret: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth",
		// this will expire the cookie
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
}
