package main_test

import (
	"reflect"
	"testing"

	. "github.com/cdrpl/idlemon-server"
)

func TestSignUpReqTags(t *testing.T) {
	req := SignUpReq{Name: "", Email: "", Pass: "password"}
	field, _ := reflect.TypeOf(req).FieldByName("Name")

	expected := `json:"name" validate:"required,alphanum,min=2,max=16"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}

	field, _ = reflect.TypeOf(req).FieldByName("Email")
	expected = `json:"email" validate:"required,email,max=255"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}

	field, _ = reflect.TypeOf(req).FieldByName("Pass")
	expected = `json:"pass" validate:"required,min=8,max=255"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}
}

func TestSignUpReqSanitize(t *testing.T) {
	req := SignUpReq{Name: " DoGName  ", Email: "  test@eMaIl.CoM  "}

	req.Sanitize()

	expected := "DoGName"
	if req.Name != expected {
		t.Errorf("name is not properly sanitized, expected: %v, received: %v", expected, req.Name)
	}

	expected = "test@email.com"
	if req.Email != expected {
		t.Errorf("email is not properly sanitized, expected: %v, received: %v", expected, req.Email)
	}
}

func TestSignInReqTags(t *testing.T) {
	req := SignInReq{Email: "", Pass: "password"}

	field, _ := reflect.TypeOf(req).FieldByName("Email")
	expected := `json:"email" validate:"required,email,max=255"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}

	field, _ = reflect.TypeOf(req).FieldByName("Pass")
	expected = `json:"pass" validate:"required,min=8,max=255"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}
}

func TestSignInReqSanitize(t *testing.T) {
	req := SignInReq{Email: "  test@eMaIl.CoM  "}

	req.Sanitize()

	expected := "test@email.com"
	if req.Email != expected {
		t.Errorf("email is not properly sanitized, expected: %v, received: %v", expected, req.Email)
	}
}

func TestUserRenameReqTags(t *testing.T) {
	req := UserRenameReq{Name: ""}

	field, _ := reflect.TypeOf(req).FieldByName("Name")
	expected := `json:"name" validate:"required,alphanum,min=2,max=16"`
	if string(field.Tag) != expected {
		t.Errorf("expected name tag: %v, received: %v", expected, field.Tag)
	}
}

func TestUserRenameReqSanitize(t *testing.T) {
	req := UserRenameReq{Name: "    dogshit        "}

	req.Sanitize()

	if req.Name != "dogshit" {
		t.Errorf("expected name to equal dogshit, received: %v", req.Name)
	}
}
