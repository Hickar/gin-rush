package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/mailer"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestCreateUser(t *testing.T) {
	r := gin.Default()
	r.POST("/api/user", CreateUser)

	mailer.Setup()
	db, mock := database.NewMockDB()
	defer db.Close()

	tests := []struct {
		Name         string
		BodyData     models.User
		ExpectedCode int
		Msg          string
	}{
		{
			"Create New User",
			models.User{Name: "NewUser", Email: "hickar@icloud.com", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusCreated,
			"Should return 201, this is new not existent user",
		},
		{
			"Create Existing User",
			models.User{Name: "NewUser", Email: "hickar@icloud.com", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusCreated,
			"Should return 409, user with these credentials already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			user := tt.BodyData

			mock.ExpectBegin()
			mock.ExpectExec("INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`password`,`salt`,`bio`,`avatar`,`birth_date`,`enabled`,`confirmation_code`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").
				WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Email, user.Password, user.Salt, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, user.ConfirmationCode).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			err := db.Create(&tt.BodyData)
			if err != nil {
				t.Errorf("Error during user creation: %s", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB query expectations were not met: %s", err)
			}
		})
	}
}

func TestUserCredentialsValidators(t *testing.T) {
	r := gin.Default()
	r.POST("/api/user", CreateUser)

	tests := []struct {
		Name         string
		BodyData     request.CreateUserRequest
		ExpectedCode int
		Msg          string
	}{
		{
			"Blank name",
			request.CreateUserRequest{Name: "", Email: "dummy@email.io", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, name is blank",
		},
		{
			"Blank email",
			request.CreateUserRequest{Name: "someUser", Email: "", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, email is blank",
		},
		{
			"Invalid email",
			request.CreateUserRequest{Name: "someUser", Email: "invalid.email", Password: "Pass/w0rd"},
			http.StatusUnprocessableEntity,
			"Should return 422, specified email doesn't meet requirements of RFC 3696 standard",
		},
		{
			"Invalid password",
			request.CreateUserRequest{Name: "someUser", Email: "dummy@email.io", Password: "v"},
			http.StatusUnprocessableEntity,
			"Must return 422, password must be length of 8, contain one uppercase character, symbol and digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			bodyData, _ := json.Marshal(tt.BodyData)
			buf := bytes.NewBuffer(bodyData)

			req, err := http.NewRequest("POST", "/api/user", buf)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				t.Fatalf("Error during request construction: %s", err)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Returned code %d – %s", w.Code, tt.Msg)
			}
		})
	}
}

func TestMain(m *testing.M) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", security.NotBlank)
		v.RegisterValidation("validemail", security.ValidEmail)
		v.RegisterValidation("validpassword", security.ValidPassword)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}