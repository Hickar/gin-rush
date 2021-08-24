package security

import (
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var ValidEmail validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(string)
	_, err := mail.ParseAddress(field)

	return err == nil
}

var ValidPassword validator.Func = func(fl validator.FieldLevel) bool {
	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
		hasSymbol bool
	)
	field := fl.Field().Interface().(string)

	if len(field) < 6 || len(field) > 64 {
		return false
	}

	for _, c := range field {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSymbol = true
		case unicode.IsLetter(c):
			hasLower = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSymbol && !IsBlank(field)
}

var NotBlank validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(string)
	test := !IsBlank(field)
	return test
}

var ValidBirthDate validator.Func = func(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(string)
	date, err := time.Parse("2006-01-02", field)

	if err != nil {
		return false
	}

	now := time.Now()
	return !date.After(now)
}

func IsBlank(str string) bool {
	return strings.Trim(str, " \t\n") == ""
}