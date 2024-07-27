package main

import (
	"context"
	"net/http"
)

type authCtxType string

var AUTH_CTX_KEY = authCtxType("req_auth")

type RequestUser struct {
	Id       string
	Username string
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secretCookie, err := r.Cookie("auth")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userId, err := rc.HGet(r.Context(), "auths", secretCookie.Value).Result()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		fields, err := rc.HMGet(r.Context(), "user:"+userId, "auth", "username").Result()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if fields[0] != secretCookie.Value {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AUTH_CTX_KEY, &RequestUser{
			Id:       userId,
			Username: fields[1].(string),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getRequestUser(ctx context.Context) *RequestUser {
	user, ok := ctx.Value(AUTH_CTX_KEY).(*RequestUser)

	if !ok {
		return nil
	}

	return user
}
