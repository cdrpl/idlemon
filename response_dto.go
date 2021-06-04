package main

// Sign in response
type SignInRes struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
