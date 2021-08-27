package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/internal/validators"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/mailer"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func TestCreateUser(t *testing.T) {
	r := gin.New()
	r.POST("/api/user", CreateUser)

	mailer.Setup()
	db, mock := database.NewMockDB()
	defer db.Close()

	tests := []struct {
		Name         string
		BodyData     models.User
		ExpectedCode int
		MockSetup    func(mock sqlmock.Sqlmock, user models.User)
		Msg          string
	}{
		{
			"Success",
			models.User{Name: "NewUser", Email: "dummy@email.io", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusCreated,
			func(mock sqlmock.Sqlmock, user models.User) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT(.*)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			"Should return 201, this is new user",
		},
		{
			"Existing",
			models.User{Name: "NewUser", Email: "dummy@email.io", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusConflict,
			func(mock sqlmock.Sqlmock, user models.User) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT(.*)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, sqlmock.AnyArg()).
					WillReturnError(errors.New("some err"))
				mock.ExpectRollback()
			},
			"Should return 409, user with these credentials already exists",
		},
		{
			"Invalid",
			models.User{Name: "", Email: "dummy@email.io", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusUnprocessableEntity,
			func(mock sqlmock.Sqlmock, user models.User) {},
			"Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			bodyData, _ := json.Marshal(tt.BodyData)
			buf := bytes.NewBuffer(bodyData)

			tt.MockSetup(mock, tt.BodyData)

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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB query expectations were not met: %s", err)
			}
		})
	}
}

func TestAuthorizeUser(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.POST("/api/authorize", AuthorizeUser)

	tests := []struct {
		Name           string
		BodyData       request.AuthUserRequest
		ActualPassword string
		ExpectedCode   int
		MockSetup      func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte)
		Msg            string
	}{
		{
			Name:           "Success",
			BodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			ActualPassword: "Pass/w0rd",
			ExpectedCode:   http.StatusOK,
			MockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", email, actualPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)
			},
			Msg: "Should return 200, user credentials are correct",
		},
		{
			Name:           "WrongPassword",
			BodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			ActualPassword: "notGuessedPassword",
			ExpectedCode:   http.StatusConflict,
			MockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", email, actualPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)
			},
			Msg: "Should return 409, user credentials are incorrect",
		},
		{
			Name:           "NotFound",
			BodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			ActualPassword: "",
			ExpectedCode:   http.StatusNotFound,
			MockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(email).WillReturnError(gorm.ErrRecordNotFound)
			},
			Msg: "Should return 404, user doesn't exist",
		},
		{
			Name:           "Invalid",
			BodyData:       request.AuthUserRequest{Email: "", Password: "Pass/w0rd"},
			ActualPassword: "",
			ExpectedCode:   http.StatusUnprocessableEntity,
			MockSetup:      func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {},
			Msg:            "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reqEnc, _ := json.Marshal(tt.BodyData)
			reqByte := bytes.NewBuffer(reqEnc)

			salt, _ := security.RandomBytes(16)
			hashedPassword, _ := security.HashPassword(tt.ActualPassword, salt)

			tt.MockSetup(mock, tt.BodyData.Email, hashedPassword, salt)

			req, _ := http.NewRequest("POST", "/api/authorize", reqByte)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Expected code %d, got %d instead", tt.ExpectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB query expectations were not met: %s", err)
			}
		})
	}
}

func jwtMiddlewareMock(id uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", id)
		c.Next()
	}
}

func TestUpdateUser(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.PATCH("/api/user", UpdateUser)

	tests := []struct {
		Name         string
		BodyData     request.UpdateUserRequest
		GuessID      int
		ActualID     int
		ExpectedCode int
		MockSetup    func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int)
		Msg          string
	}{
		{
			Name:         "Success",
			BodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusNoContent,
			MockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(guessID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)

				// Update User
				formattedTime, _ := time.Parse("2006-01-02", reqData.BirthDate)
				query = "UPDATE(.*)"
				mock.ExpectBegin()
				mock.ExpectExec(query).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), reqData.Name, "dummy@email.io", []byte("pass"), []byte("salt"), reqData.Bio, reqData.Avatar, formattedTime, false, "code", guessID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Msg: "Should return 204, request was sent from same user it's must update",
		},
		{
			Name:         "NotFound",
			BodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusNotFound,
			MockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			Msg: "Should return 404, user doesn't exist",
		},
		{
			Name:         "Forbidden",
			BodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			GuessID:      trueID,
			ActualID:     43,
			ExpectedCode: http.StatusForbidden,
			MockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(actualID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)
			},
			Msg: "Should return 403, request was made by other user",
		},
		{
			Name:         "Invalid",
			BodyData:     request.UpdateUserRequest{Name: "", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: ""},
			GuessID:      trueID,
			ActualID:     43,
			ExpectedCode: http.StatusUnprocessableEntity,
			MockSetup:    func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {},
			Msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.BodyData)
			reqBytes := bytes.NewBuffer(reqBody)

			req, _ := http.NewRequest("PATCH", "/api/user", reqBytes)
			w := httptest.NewRecorder()

			tt.MockSetup(mock, tt.BodyData, tt.GuessID, tt.ActualID)

			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Expected code %d, got %d", tt.ExpectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.GET("/api/user/:id", GetUser)

	tests := []struct {
		Name         string
		GuessID      int
		ActualID     interface{}
		ExpectedCode interface{}
		MockSetup    func(mock sqlmock.Sqlmock, guessID, actualID interface{})
		Msg          string
	}{
		{
			Name:         "Success",
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusOK,
			MockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(guessID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)
			},
			Msg: "Should return 200, request was sent from same user it's must update",
		},
		{
			Name:         "NotFound",
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusNotFound,
			MockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			Msg: "Should return 404, user doesn't exist",
		},
		{
			Name:         "Forbidden",
			GuessID:      trueID,
			ActualID:     43,
			ExpectedCode: http.StatusForbidden,
			MockSetup:    func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			Msg:          "Should return 403, request was made by other user",
		},
		{
			Name:         "Invalid",
			GuessID:      trueID,
			ActualID:     "notValidID",
			ExpectedCode: http.StatusUnprocessableEntity,
			MockSetup:    func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			Msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockSetup(mock, tt.GuessID, tt.ActualID)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/user/%v", tt.ActualID), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Expected code %d, got %d", tt.ExpectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.DELETE("/api/user/:id", DeleteUser)

	tests := []struct {
		Name         string
		GuessID      interface{}
		ActualID     interface{}
		ExpectedCode int
		MockSetup    func(mock sqlmock.Sqlmock, guessID, actualID interface{})
		Msg          string
	}{
		{
			Name:         "Success",
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusNoContent,
			MockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(guessID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE(.*)").
					WithArgs(sqlmock.AnyArg(), guessID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Msg: "Should return 204, request was sent from same user it's must delete",
		},
		{
			Name:         "NotFound",
			GuessID:      trueID,
			ActualID:     trueID,
			ExpectedCode: http.StatusNotFound,
			MockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			Msg: "Should return 404, user doesn't exist",
		},
		{
			Name:         "Forbidden",
			GuessID:      trueID,
			ActualID:     43,
			ExpectedCode: http.StatusForbidden,
			MockSetup:    func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			Msg:          "Should return 403, request was made by other user",
		},
		{
			Name:         "Invalid",
			GuessID:      trueID,
			ActualID:     "notValidID",
			ExpectedCode: http.StatusUnprocessableEntity,
			MockSetup:    func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			Msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockSetup(mock, tt.GuessID, tt.ActualID)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/user/%v", tt.ActualID), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Expected code %d, got %d", tt.ExpectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestEnableUser(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()

	userID := 42
	trueCode := utils.RandomString(30)

	r.Use(jwtMiddlewareMock(uint(userID)))
	r.GET("/api/authorize/email/challenge/:code", EnableUser)

	tests := []struct {
		Name          string
		ChallengeCode interface{}
		ExpectedCode  int
		MockSetup     func(mock sqlmock.Sqlmock, userID, code interface{})
		Msg           string
	}{
		{
			Name:          "Success",
			ChallengeCode: trueCode,
			ExpectedCode:  http.StatusOK,
			MockSetup: func(mock sqlmock.Sqlmock, userID, code interface{}) {
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(userID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, code)
				mock.ExpectQuery("SELECT(.*)").WithArgs(code).WillReturnRows(rows)

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE(.*)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, true, code, userID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			Msg: "Should return 200, request was sent from same user it's must delete",
		},
		{
			Name:          "NotFound",
			ChallengeCode: trueCode,
			ExpectedCode:  http.StatusNotFound,
			MockSetup: func(mock sqlmock.Sqlmock, userID, code interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(code).WillReturnError(gorm.ErrRecordNotFound)
			},
			Msg: "Should return 404, user doesn't exist",
		},
		{
			Name:          "Invalid",
			ChallengeCode: "invalidCode",
			ExpectedCode:  http.StatusUnprocessableEntity,
			MockSetup:     func(mock sqlmock.Sqlmock, userID, code interface{}) {},
			Msg:           "Should return 422, verification code is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.MockSetup(mock, userID, tt.ChallengeCode)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/authorize/email/challenge/%v", tt.ChallengeCode), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.ExpectedCode {
				t.Errorf("Expected code %d, got %d", tt.ExpectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", validators.NotBlank)
		v.RegisterValidation("validemail", validators.ValidEmail)
		v.RegisterValidation("validpassword", validators.ValidPassword)
		v.RegisterValidation("validbirthdate", validators.ValidBirthDate)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}
