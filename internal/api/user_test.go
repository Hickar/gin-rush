package api
//
//import (
//	"bytes"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/Hickar/gin-rush/internal/repositories"
//	"github.com/Hickar/gin-rush/internal/security"
//	"github.com/Hickar/gin-rush/pkg/database"
//	"github.com/gin-gonic/gin"
//	"github.com/gin-gonic/gin/binding"
//	"github.com/go-playground/validator/v10"
//)
//
//func TestCreateUser(t *testing.T) {
//	r := gin.Default()
//	r.POST("/api/user", CreateUser)
//
//	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//		v.RegisterValidation("notblank", security.NotBlank)
//		v.RegisterValidation("validemail", security.ValidEmail)
//		v.RegisterValidation("validpassword", security.ValidPassword)
//	}
//
//	database.Setup()
//
//	tests := []struct{
//		BodyData string
//		ExpectedCode int
//		Msg string
//	}{
//		{
//			`{"name":"","email":"dummy@email.io","password":"Pass/w0rd"}`,
//			http.StatusUnprocessableEntity,
//			"Should return 422, name is blank",
//		},
//		{
//			`{"name":"someUser","email":"","password":"Pass/w0rd"}`,
//			http.StatusUnprocessableEntity,
//			"Should return 422, email is blank",
//		},
//		{
//			`{"name":"someUser","email":"wrong.email","password":"Pass/w0rd"}`,
//			http.StatusUnprocessableEntity,
//			"Should return 422, specified email doesn't meet requirements of RFC 3696 standard",
//		},
//		{
//			`{"name":"someUser","email":"dummy@email.io","password":"v"}`,
//			http.StatusUnprocessableEntity,
//			"Must return 422, password must be length of 8, contain one uppercase character, symbol and digit",
//		},
//		{
//			`{"name":"SampleUser","email":"dummy@email.io","password":"Pass/w0rd"}`,
//			http.StatusCreated,
//			"Must return 401, registration must have gone successfully",
//		},
//		{
//			`{"name":"SampleUser","email":"dummy@email.io","password":"Pass/w0rd"}`,
//			http.StatusConflict,
//			"Must return 409, user with such credentials already exists",
//		},
//	}
//
//	for _, testData := range tests {
//		bodyData := testData.BodyData
//		req, err := http.NewRequest("POST", "/api/user", bytes.NewBufferString(bodyData))
//		req.Header.Set("Content-Type", "application/json")
//
//		if err != nil {
//			t.Errorf("Error during request construction: %s", err)
//		}
//
//		w := httptest.NewRecorder()
//		r.ServeHTTP(w, req)
//
//		if w.Code != testData.ExpectedCode {
//			repositories.DB.Where("email = ?", "dummy@email.io").Delete(&repositories.User{})
//			t.Errorf("Returned code %d – %s", w.Code, testData.Msg)
//		}
//	}
//}