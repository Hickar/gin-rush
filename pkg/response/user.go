package response

type AuthUserResponse struct {
	Token string `json:"token"`
}

type UpdateUserResponse struct {
	Name      string `json:"name" binding:"required,max=128,notblank" maxLength:"128"`
	Bio       string `json:"bio" binding:"max=512" maxLength:"512"`
	Avatar    string `json:"avatar"`
	BirthDate string `json:"birth_date" binding:"validbirthdate"`
}