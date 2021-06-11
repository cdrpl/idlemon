package main

// Sign in response
type SignInRes struct {
	Token         string         `json:"token"`
	User          User           `json:"user"`
	Units         []Unit         `json:"units"`
	UserResources []UserResource `json:"resources"`
}
