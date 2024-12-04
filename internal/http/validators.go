package http

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var emailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	FieldErrors map[string]string
	FormErrors  []string
}

func (v Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) NotBlank(key, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddFieldError(key, fmt.Sprintf("%s cannot be blank", key))
	}
}

func (v *Validator) NotNegative(key string, value float32) {
	if value < 0 {
		v.AddFieldError(key, fmt.Sprintf("%s cannot be less than zero", key))
	}
}

func (v *Validator) MaxChars(key, value string, n int) {
	if utf8.RuneCountInString(value) > n {
		v.AddFieldError(key, fmt.Sprintf("%s cannot be more than %d characters long", key, n))
	}
}

func (v *Validator) MinChars(key, value string, n int) {
	if utf8.RuneCountInString(value) < n {
		v.AddFieldError(key, fmt.Sprintf("%s field must be at least %d characters long", key, n))
	}
}

func (v *Validator) ValidEmail(key, value string) {
	if !emailRX.MatchString(value) {
		v.AddFieldError(key, "This field must be a valid email address")
	}
}

func (v *Validator) AddFormError(message string) {
	v.FormErrors = append(v.FormErrors, message)
}
