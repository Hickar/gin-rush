package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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
			"Create New User",
			models.User{Name: "NewUser", Email: "dummy@email.io", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusCreated,
			func(mock sqlmock.Sqlmock, user models.User) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`password`,`salt`,`bio`,`avatar`,`birth_date`,`enabled`,`confirmation_code`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			"Should return 201, this is new user",
		},
		{
			"Create Existing User",
			models.User{Name: "NewUser", Email: "dummy@email.io", Password: []byte("Test/P4ass"), Salt: []byte("salt"), ConfirmationCode: "verificationCode"},
			http.StatusConflict,
			func(mock sqlmock.Sqlmock, user models.User) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`password`,`salt`,`bio`,`avatar`,`birth_date`,`enabled`,`confirmation_code`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), user.Name, user.Email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false, sqlmock.AnyArg()).
					WillReturnError(errors.New("some err"))
				mock.ExpectRollback()
			},
			"Should return 409, user with these credentials already exists",
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

func TestAuthorizeUserSuccess(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.POST("/api/authorize", AuthorizeUser)

	reqData := request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"}
	reqEnc, _ := json.Marshal(reqData)
	reqByte := bytes.NewBuffer(reqEnc)

	salt, _ := security.RandomBytes(16)
	hashedPassword, _ := security.HashPassword(reqData.Password, salt)

	req, _ := http.NewRequest("POST", "/api/authorize", reqByte)
	w := httptest.NewRecorder()

	query := "SELECT * FROM `users` WHERE email = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
	rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", reqData.Email, hashedPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")

	mock.ExpectQuery(query).WithArgs(reqData.Email).WillReturnRows(rows)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected code %d, got %d instead", http.StatusOK, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB query expectations were not met: %s", err)
	}
}

func TestAuthorizeUserWrongPassword(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.POST("/api/authorize", AuthorizeUser)

	reqData := request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"}
	reqEnc, _ := json.Marshal(reqData)
	reqByte := bytes.NewBuffer(reqEnc)

	salt, _ := security.RandomBytes(16)
	hashedPassword, _ := security.HashPassword("WrongPassword", salt)

	req, _ := http.NewRequest("POST", "/api/authorize", reqByte)
	w := httptest.NewRecorder()

	query := "SELECT * FROM `users` WHERE email = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
	rows := sqlmock.NewRows(columns).AddRow(0, time.Now(), time.Now(), sql.NullTime{}, "NewUser", reqData.Email, hashedPassword, salt, sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")

	mock.ExpectQuery(query).WithArgs(reqData.Email).WillReturnRows(rows)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected code %d, got %d instead", http.StatusConflict, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB query expectations were not met: %s", err)
	}
}

func TestAuthorizeUserNotFound(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.POST("/api/authorize", AuthorizeUser)

	reqData := request.AuthUserRequest{Email: "dummy@email.io", Password: "Pass/w0rd"}
	reqEnc, _ := json.Marshal(reqData)
	reqByte := bytes.NewBuffer(reqEnc)

	req, _ := http.NewRequest("POST", "/api/authorize", reqByte)
	w := httptest.NewRecorder()

	query := "SELECT * FROM `users` WHERE email = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	mock.ExpectQuery(query).WithArgs(reqData.Email).WillReturnError(gorm.ErrRecordNotFound)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected code %d, got %d instead", http.StatusNotFound, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB query expectations were not met: %s", err)
	}
}

func jwtMiddlewareMock() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	}
}

func TestUpdateUserSuccess(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.Use(jwtMiddlewareMock())
	r.PATCH("/api/user", UpdateUser).Use(jwtMiddlewareMock())

	reqData := &request.UpdateUserRequest{
		Name:      "NewUser2",
		Bio:       "Hi, it's info about me",
		Avatar:    "https://some.img.service/myPhotoId",
		BirthDate: "1989-04-19",
	}
	reqBody, _ := json.Marshal(reqData)
	reqBytes := bytes.NewBuffer(reqBody)

	req, _ := http.NewRequest("PATCH", "/api/user", reqBytes)
	w := httptest.NewRecorder()

	// Find User
	query := "SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
	rows := sqlmock.NewRows(columns).AddRow(42, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
	mock.ExpectQuery(query).WithArgs(42).WillReturnRows(rows)

	// Update User
	formattedTime, _ := time.Parse("2006-01-02", reqData.BirthDate)
	query = "UPDATE `users` SET `created_at`=?,`updated_at`=?,`deleted_at`=?,`name`=?,`email`=?,`password`=?,`salt`=?,`bio`=?,`avatar`=?,`birth_date`=?,`enabled`=?,`confirmation_code`=? WHERE `id` = ?"
	mock.ExpectBegin()
	mock.ExpectExec(query).WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), reqData.Name, "dummy@email.io", []byte("pass"), []byte("salt"), reqData.Bio, reqData.Avatar, formattedTime, false, "code", 42).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected code %d, got %d", http.StatusNoContent, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB expectations were not met: %s", err)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.Use(jwtMiddlewareMock())
	r.PATCH("/api/user", UpdateUser).Use(jwtMiddlewareMock())

	reqData := &request.UpdateUserRequest{
		Name:      "NewUser2",
		Bio:       "Hi, it's info about me",
		Avatar:    "https://some.img.service/myPhotoId",
		BirthDate: "1989-04-19",
	}
	reqBody, _ := json.Marshal(reqData)
	reqBytes := bytes.NewBuffer(reqBody)

	req, _ := http.NewRequest("PATCH", "/api/user", reqBytes)
	w := httptest.NewRecorder()

	// Find User
	query := "SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	mock.ExpectQuery(query).WithArgs(42).WillReturnError(gorm.ErrRecordNotFound)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected code %d, got %d", http.StatusNotFound, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB expectations were not met: %s", err)
	}
}

func TestUpdateUserForbidden(t *testing.T) {
	_, mock := database.NewMockDB()
	r := gin.New()
	r.Use(jwtMiddlewareMock())
	r.PATCH("/api/user", UpdateUser).Use(jwtMiddlewareMock())

	reqData := &request.UpdateUserRequest{
		Name:      "NewUser2",
		Bio:       "Hi, it's info about me",
		Avatar:    "https://some.img.service/myPhotoId",
		BirthDate: "1989-04-19",
	}
	reqBody, _ := json.Marshal(reqData)
	reqBytes := bytes.NewBuffer(reqBody)

	req, _ := http.NewRequest("PATCH", "/api/user", reqBytes)
	w := httptest.NewRecorder()

	// Find User
	query := "SELECT * FROM `users` WHERE `users`.`id` = ? AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1"
	columns := []string{"id", "created_at", "updated_at", "deleted_at", "name", "email", "password", "salt", "bio", "avatar", "birth_date", "enabled", "confirmation_code"}
	rows := sqlmock.NewRows(columns).AddRow(43, time.Now(), time.Now(), sql.NullTime{}, "NewUser", "dummy@email.io", []byte("pass"), []byte("salt"), sql.NullString{}, sql.NullString{}, sql.NullTime{}, false, "code")
	mock.ExpectQuery(query).WithArgs(42).WillReturnRows(rows)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected code %d, got %d", http.StatusForbidden, w.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some of DB expectations were not met: %s", err)
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