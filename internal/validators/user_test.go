package validators

import (
	"testing"

	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestUserValidators(t *testing.T) {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("unexpected error during validator initialization")
	}

	v.RegisterValidation("notblank", NotBlank)
	v.RegisterValidation("validemail", ValidEmail)
	v.RegisterValidation("validpassword", ValidPassword)
	v.RegisterValidation("validbirthdate", ValidBirthDate)

	v.Struct(request.CreateUserRequest{Name: "someUser", Email: "invalid.email", Password: "Pass/w0rd"})

	tests := []struct {
		Name      string
		Data      interface{}
		ShouldErr bool
		Msg       string
	}{
		{
			"BlankNameShouldFail",
			request.CreateUserRequest{Name: "", Email: "dummy@email.io", Password: "Pass/w0rd"},
			true,
			"Name is blank",
		},
		{
			"BlankNameShouldPass",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "Pass/w0rd"},
			false,
			"Name is not blank",
		},
		{
			"BlankEmailShouldFail",
			request.CreateUserRequest{Name: "someUser", Email: "", Password: "Pass/w0rd"},
			true,
			"Email is blank",
		},
		{
			"BlankEmailShouldPass",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "Pass/w0rd"},
			false,
			"Email is not blank",
		},
		{
			"InvalidEmailShouldFail",
			request.CreateUserRequest{Name: "someUser", Email: "invalid.email", Password: "Pass/w0rd"},
			true,
			"Specified email DOESN'T MEET requirements of RFC 3696 standard",
		},
		{
			"InvalidEmailShouldPass",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "Pass/w0rd"},
			false,
			"Specified email MEETS requirements of RFC 3696 standard",
		},
		{
			"InvalidPasswordShouldFail",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "v"},
			true,
			"Password must be length of 8, contain one uppercase character, symbol and digit",
		},
		{
			"InvalidPasswordShouldPass",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "Pass/w0rd"},
			false,
			"Password must be length of 8, contain one uppercase character, symbol and digit",
		},
		{
			"InvalidBirthdateShouldFail",
			request.UpdateUserRequest{Name: "someUser", BirthDate: "2077-01-01"},
			true,
			"Password must be length of 8, contain one uppercase character, symbol and digit",
		},
		{
			"InvalidBirthdateShouldPass",
			request.UpdateUserRequest{Name: "someUser", BirthDate: "1989-01-01"},
			false,
			"Password must be length of 8, contain one uppercase character, symbol and digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := v.Struct(tt.Data)

			if (err != nil && !tt.ShouldErr) || (err == nil && tt.ShouldErr) {
				t.Errorf("Error: %s", tt.Msg)
			}
		})
	}
}
