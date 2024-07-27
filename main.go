package main

import (
	"net/http"

	"github.com/redis/go-redis/v9"
)

var rc *redis.Client

func main() {
	mux := http.NewServeMux()

	rc = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	mux.HandleFunc("/register", register)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/logout", authMiddleware(logout))

	mux.HandleFunc("/post", authMiddleware(post))
	mux.HandleFunc("/timeline", authMiddleware(timeline))
	mux.HandleFunc("/newusers", authMiddleware(newusers))
	mux.HandleFunc("/profile", authMiddleware(profile))
	mux.HandleFunc("/follow", authMiddleware(follow))

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
