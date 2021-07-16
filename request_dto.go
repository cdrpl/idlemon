package main

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

type RequestDTO interface {
	Validate() error
}

type ChatMessageSendReq struct {
	Message string `json:"message"`
}

func (r *ChatMessageSendReq) Validate() error {
	r.Message = strings.TrimSpace(r.Message)

	// validate message minimum length
	if len(r.Message) < CHAT_MESSAGE_MIN_LEN {
		return fmt.Errorf("message must have at least %v characters", CHAT_MESSAGE_MIN_LEN)
	}

	// validate message max length
	if len(r.Message) > CHAT_MESSAGE_MAX_LEN {
		return fmt.Errorf("message cannot have more than %v characters", CHAT_MESSAGE_MAX_LEN)
	}

	return nil
}

type SignUpReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Pass  string `json:"pass"`
}

func (r *SignUpReq) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.TrimSpace(r.Email)
	r.Email = strings.ToLower(r.Email)

	if err := ValidateUserName(r.Name); err != nil {
		return err
	}

	if err := ValidateEmail(r.Email); err != nil {
		return err
	}

	if err := ValidateUserPass(r.Pass); err != nil {
		return err
	}

	return nil
}

// Sign in request
type SignInReq struct {
	Email string `json:"email"`
	Pass  string `json:"pass"`
}

func (r *SignInReq) Validate() error {
	r.Email = strings.TrimSpace(r.Email)
	r.Email = strings.ToLower(r.Email)

	if err := ValidateEmail(r.Email); err != nil {
		return err
	}

	if err := ValidateUserPass(r.Pass); err != nil {
		return err
	}

	return nil
}

type UserRenameReq struct {
	Name string `json:"name"`
}

func (r *UserRenameReq) Validate() error {
	r.Name = strings.TrimSpace(r.Name)

	if err := ValidateUserName(r.Name); err != nil {
		return err
	}

	return nil
}

func ValidateUserName(name string) error {
	if len(name) < USER_NAME_MIN {
		return fmt.Errorf("user name must have at least %v characters", USER_NAME_MIN)
	}

	if len(name) > USER_NAME_MAX {
		return fmt.Errorf("user name cannot have more than %v characters", USER_NAME_MAX)
	}

	return nil
}

func ValidateEmail(email string) error {
	if len(email) > USER_EMAIL_MAX {
		return fmt.Errorf("email cannot have more than %v characters", USER_EMAIL_MAX)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("email address is invalid")
	}

	return nil
}

func ValidateUserPass(pass string) error {
	if len(pass) < USER_PASS_MIN {
		return fmt.Errorf("user password must have at least %v characters", USER_PASS_MIN)
	}

	if len(pass) > USER_PASS_MAX {
		return fmt.Errorf("user password cannot have more than %v characters", USER_PASS_MAX)
	}

	return nil
}
