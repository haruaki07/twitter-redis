package main

type Post struct {
	Id       string `json:"id"`
	Body     string `json:"body"`
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Time     string `json:"time"`
}
