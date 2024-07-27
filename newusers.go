package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type NewUser struct {
	Username string `json:"username"`
	Time     string `json:"time"`
}

func newusers(w http.ResponseWriter, r *http.Request) {
	// get 10 latest registered users
	res, err := rc.ZRevRangeWithScores(r.Context(), "users_by_time", 0, 9).Result()
	if err != nil {
		log.Printf("failed to get latest registered users: %e\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	users := make([]*NewUser, 0, 10)
	for _, z := range res {
		user := &NewUser{}
		user.Username = z.Member.(string)
		user.Time = fmt.Sprintf("%9.f", z.Score)
		users = append(users, user)
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(users)
}
