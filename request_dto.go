package main

import (
	"strings"
)

type RequestDTO interface {
	Sanitize()
}

type SignUpReq struct {
	Name  string `json:"name" validate:"required,alphanum,min=2,max=16"`
	Email string `json:"email" validate:"required,email,max=255"`
	Pass  string `json:"pass" validate:"required,min=8,max=255"`
}

func (r *SignUpReq) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(r.Email)
	r.Email = strings.ToLower(r.Email)
}

// Sign in request
type SignInReq struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Pass  string `json:"pass" validate:"required,min=8,max=255"`
}

func (s *SignInReq) Sanitize() {
	s.Email = strings.TrimSpace(s.Email)
	s.Email = strings.ToLower(s.Email)
}

type UserRenameReq struct {
	Name string `json:"name" validate:"required,alphanum,min=2,max=16"`
}

func (r *UserRenameReq) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
}
