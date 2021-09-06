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
	"github.com/Hickar/gin-rush/internal/cache"
	"github.com/Hickar/gin-rush/internal/config"
	"github.com/Hickar/gin-rush/internal/models"
	"github.com/Hickar/gin-rush/internal/security"
	"github.com/Hickar/gin-rush/internal/validators"
	"github.com/Hickar/gin-rush/pkg/database"
	"github.com/Hickar/gin-rush/pkg/response"

	//"github.com/Hickar/gin-rush/pkg/mailer"
	"github.com/Hickar/gin-rush/pkg/request"
	"github.com/Hickar/gin-rush/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redismock/v8"
	"gorm.io/gorm"
)

func TestCreateUser(t *testing.T) {
	r := gin.New()
	r.POST("/api/user", CreateUser)

	//conf := config.GetConfig()
	//mailer.NewMailer(&conf.Gmail)
	db, mock := database.NewMockDB()
	defer db.Close()

	tests := []struct {
		name         string
		bodyData     models.User
		expectedCode int
		dbMockSetup  func(mock sqlmock.Sqlmock, user models.User)
		msg          string
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
		t.Run(tt.name, func(t *testing.T) {
			bodyData, _ := json.Marshal(tt.bodyData)
			buf := bytes.NewBuffer(bodyData)

			tt.dbMockSetup(mock, tt.bodyData)

			req, err := http.NewRequest("POST", "/api/user", buf)
			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				t.Fatalf("Error during request construction: %s", err)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Returned code %d – %s", w.Code, tt.msg)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB query expectations were not met: %s", err)
			}
		})
	}
}

func TestAuthorizeUser(t *testing.T) {
	_, dbMock := database.NewMockDB()
	r := gin.New()
	r.POST("/api/authorize", AuthorizeUser)

	tests := []struct {
		name           string
		bodyData       request.AuthUserRequest
		actualPassword string
		expectedCode   int
		dbMockSetup    func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte)
		msg            string
	}{
		{
			name:           "Success",
			bodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			actualPassword: "Pass/w0rd",
			expectedCode:   http.StatusOK,
			dbMockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", email, actualPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)
			},
			msg: "Should return 200, user credentials are correct",
		},
		{
			name:           "WrongPassword",
			bodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			actualPassword: "notGuessedPassword",
			expectedCode:   http.StatusConflict,
			dbMockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", email, actualPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(email).WillReturnRows(rows)
			},
			msg: "Should return 409, user credentials are incorrect",
		},
		{
			name:           "NotFound",
			bodyData:       request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"},
			actualPassword: "",
			expectedCode:   http.StatusNotFound,
			dbMockSetup: func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(email).WillReturnError(gorm.ErrRecordNotFound)
			},
			msg: "Should return 404, user doesn't exist",
		},
		{
			name:           "Invalid",
			bodyData:       request.AuthUserRequest{Email: "", Password: "Pass/w0rd"},
			actualPassword: "",
			expectedCode:   http.StatusUnprocessableEntity,
			dbMockSetup:    func(mock sqlmock.Sqlmock, email string, actualPassword, salt []byte) {},
			msg:            "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqEnc, _ := json.Marshal(tt.bodyData)
			reqByte := bytes.NewBuffer(reqEnc)

			salt, _ := security.RandomBytes(16)
			hashedPassword, _ := security.HashPassword(tt.actualPassword, salt)

			tt.dbMockSetup(dbMock, tt.bodyData.Email, hashedPassword, salt)

			req, _ := http.NewRequest("POST", "/api/authorize", reqByte)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d instead", tt.expectedCode, w.Code)
			}

			if err := dbMock.ExpectationsWereMet(); err != nil {
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
	_, dbMock := database.NewMockDB()
	_, redisMock := cache.NewCacheMock()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.PATCH("/api/user", UpdateUser)

	tests := []struct {
		name         string
		bodyData     request.UpdateUserRequest
		guessID      int
		actualID     int
		expectedCode int
		dbMockSetup  func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int)
		redisMockSetup func(mock redismock.ClientMock, guessID int)
		msg          string
	}{
		{
			name:         "Success",
			bodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusNoContent,
			dbMockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
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
			redisMockSetup: func(mock redismock.ClientMock, guessID int) {
				mock.ExpectDel(fmt.Sprintf("users:%d", guessID)).SetVal(1)
			},
			msg: "Should return 204, request was sent from same user it's must update",
		},
		{
			name:         "NotFound",
			bodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusNotFound,
			dbMockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			redisMockSetup: func(mock redismock.ClientMock, guessID int) {},
			msg: "Should return 404, user doesn't exist",
		},
		{
			name:         "Forbidden",
			bodyData:     request.UpdateUserRequest{Name: "NewUser2", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: "1989-04-19"},
			guessID:      trueID,
			actualID:     43,
			expectedCode: http.StatusForbidden,
			dbMockSetup: func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(actualID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)
			},
			redisMockSetup: func(mock redismock.ClientMock, guessID int) {},
			msg: "Should return 403, request was made by other user",
		},
		{
			name:         "Invalid",
			bodyData:     request.UpdateUserRequest{Name: "", Bio: "Hi, it's info about me", Avatar: "https://some.img.service/myPhotoId", BirthDate: ""},
			guessID:      trueID,
			actualID:     43,
			expectedCode: http.StatusUnprocessableEntity,
			dbMockSetup:  func(mock sqlmock.Sqlmock, reqData request.UpdateUserRequest, guessID, actualID int) {},
			redisMockSetup: func(mock redismock.ClientMock, guessID int) {},
			msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(tt.bodyData)
			reqBytes := bytes.NewBuffer(reqBody)

			req, _ := http.NewRequest("PATCH", "/api/user", reqBytes)
			w := httptest.NewRecorder()

			tt.dbMockSetup(dbMock, tt.bodyData, tt.guessID, tt.actualID)
			tt.redisMockSetup(redisMock, tt.guessID)

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, w.Code)
			}

			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	_, dbMock := database.NewMockDB()
	_, redisMock := cache.NewCacheMock()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.GET("/api/user/:id", GetUser)

	tests := []struct {
		name         string
		guessID      int
		actualID     interface{}
		expectedCode interface{}
		dbMockSetup  func(mock sqlmock.Sqlmock, guessID, actualID interface{})
		redisMockSetup func(mock redismock.ClientMock, guessID interface{})
		msg          string
	}{
		{
			name:         "Success",
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusOK,
			dbMockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(guessID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnRows(rows)
			},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {
				resp := response.UpdateUserResponse{
					Name:      "NewUser",
					Bio:       "",
					Avatar:    "",
					BirthDate: "0001-01-01 00:00:00 +0000 UTC",
				}
				cacheData, _ := json.Marshal(&resp)
				cacheKey := fmt.Sprintf("users:%v", guessID)
				mock.ExpectGet(cacheKey).RedisNil()
				mock.ExpectSet(cacheKey, cacheData, time.Duration(time.Hour * 168)).SetVal("")
			},
			msg: "Should return 200, request was sent from same user it's must update",
		},
		{
			name:         "NotFound",
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusNotFound,
			dbMockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {
				cacheKey := fmt.Sprintf("users:%v", guessID)
				mock.ExpectGet(cacheKey).RedisNil()
			},
			msg: "Should return 404, user doesn't exist",
		},
		{
			name:         "Forbidden",
			guessID:      trueID,
			actualID:     43,
			expectedCode: http.StatusForbidden,
			dbMockSetup:  func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {},
			msg:          "Should return 403, request was made by other user",
		},
		{
			name:         "Invalid",
			guessID:      trueID,
			actualID:     "notValidID",
			expectedCode: http.StatusUnprocessableEntity,
			dbMockSetup:  func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {},
			msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.redisMockSetup(redisMock, tt.guessID)
			tt.dbMockSetup(dbMock, tt.guessID, tt.actualID)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/user/%v", tt.actualID), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, w.Code)
			}

			if err := dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	_, dbMock := database.NewMockDB()
	_, redisMock := cache.NewCacheMock()
	r := gin.New()

	trueID := 42

	r.Use(jwtMiddlewareMock(uint(trueID)))
	r.DELETE("/api/user/:id", DeleteUser)

	tests := []struct {
		name         string
		guessID      interface{}
		actualID     interface{}
		expectedCode int
		dbMockSetup  func(mock sqlmock.Sqlmock, guessID, actualID interface{})
		redisMockSetup func(mock redismock.ClientMock, guessID interface{})
		msg          string
	}{
		{
			name:         "Success",
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusNoContent,
			dbMockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
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
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {
				mock.ExpectDel(fmt.Sprintf("users:%v", guessID)).SetVal(1)
			},
			msg: "Should return 204, request was sent from same user it's must delete",
		},
		{
			name:         "NotFound",
			guessID:      trueID,
			actualID:     trueID,
			expectedCode: http.StatusNotFound,
			dbMockSetup: func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(guessID).WillReturnError(gorm.ErrRecordNotFound)
			},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {},
			msg: "Should return 404, user doesn't exist",
		},
		{
			name:         "Forbidden",
			guessID:      trueID,
			actualID:     43,
			expectedCode: http.StatusForbidden,
			dbMockSetup:  func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {},
			msg:          "Should return 403, request was made by other user",
		},
		{
			name:         "Invalid",
			guessID:      trueID,
			actualID:     "notValidID",
			expectedCode: http.StatusUnprocessableEntity,
			dbMockSetup:  func(mock sqlmock.Sqlmock, guessID, actualID interface{}) {},
			redisMockSetup: func(mock redismock.ClientMock, guessID interface{}) {},
			msg:          "Should return 422, user credentials are invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.redisMockSetup(redisMock, tt.guessID)
			tt.dbMockSetup(dbMock, tt.guessID, tt.actualID)

			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/user/%v", tt.actualID), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, w.Code)
			}

			if err := dbMock.ExpectationsWereMet(); err != nil {
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
		name          string
		challengeCode interface{}
		expectedCode  int
		dbMockSetup   func(mock sqlmock.Sqlmock, userID, code interface{})
		msg           string
	}{
		{
			name:          "Success",
			challengeCode: trueCode,
			expectedCode:  http.StatusOK,
			dbMockSetup: func(mock sqlmock.Sqlmock, userID, code interface{}) {
				columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
				rows := sqlmock.NewRows(columns).AddRow(userID, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, code)
				mock.ExpectQuery("SELECT(.*)").WithArgs(code).WillReturnRows(rows)

				mock.ExpectBegin()
				mock.ExpectExec("UPDATE(.*)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, true, code, userID).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			msg: "Should return 200, request was sent from same user it's must delete",
		},
		{
			name:          "NotFound",
			challengeCode: trueCode,
			expectedCode:  http.StatusNotFound,
			dbMockSetup: func(mock sqlmock.Sqlmock, userID, code interface{}) {
				query := "SELECT(.*)"
				mock.ExpectQuery(query).WithArgs(code).WillReturnError(gorm.ErrRecordNotFound)
			},
			msg: "Should return 404, user doesn't exist",
		},
		{
			name:          "Invalid",
			challengeCode: "invalidCode",
			expectedCode:  http.StatusUnprocessableEntity,
			dbMockSetup:   func(mock sqlmock.Sqlmock, userID, code interface{}) {},
			msg:           "Should return 422, verification code is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.dbMockSetup(mock, userID, tt.challengeCode)

			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/authorize/email/challenge/%v", tt.challengeCode), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected code %d, got %d", tt.expectedCode, w.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Some of DB expectations were not met: %s", err)
			}
		})
	}
}

func TestMain(m *testing.M) {
	conf := config.NewConfig("../../conf/config.test.json")
	gin.SetMode(conf.Server.Mode)

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notblank", validators.NotBlank)
		v.RegisterValidation("validemail", validators.ValidEmail)
		v.RegisterValidation("validpassword", validators.ValidPassword)
		v.RegisterValidation("validbirthdate", validators.ValidBirthDate)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}
