package request

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Email    string `json:"email" binding:"required,validemail" maxLength:"128"`
	Password string `json:"password" binding:"required,validpassword" minLength:"6" maxLength:"64"`
}

type AuthUserRequest struct {
	Email    string `json:"email" binding:"required,validemail" maxLength:"128"`
	Password string `json:"password" binding:"required,validpassword" minLength:"6" maxLength:"64"`
}

type UpdateUserRequest struct {
	Name      string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Bio       string `json:"bio" binding:"max=512" maxLength:"512"`
	Avatar    string `json:"avatar"`
	BirthDate string `json:"birth_date" binding:"validbirthdate"`
}