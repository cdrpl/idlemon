package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
)

func RunStructValidator(validate *validator.Validate, v RequestDTO) (msg string, hasError bool) {
	v.Sanitize()

	err := validate.Struct(v)

	if err != nil {
		var validationErrors validator.ValidationErrors

		if errors.As(err, &validationErrors) {
			for _, err := range validationErrors {
				msg := ValidationErrMsg(err.StructField(), err.Tag(), err.Param())
				return msg, true
			}
		} else {
			log.Fatalf("struct validator error: %v\n", err)
		}
	}

	return "", false
}

// Will construct an error message based on the validation error.
func ValidationErrMsg(field string, tag string, param string) string {
	field = strings.ToLower(field)

	switch tag {
	case "required":
		return fmt.Sprintf("%v is required", field)

	case "min":
		return fmt.Sprintf("%v minimum length of %v", field, param)

	case "max":
		return fmt.Sprintf("%v maximum length of %v", field, param)

	case "email":
		return "email is invalid"

	default:
		return fmt.Sprintf("no error message for validation type %v", tag)
	}
}
