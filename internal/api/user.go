package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserDummy struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// GetUser godoc
// @Summary Get user
// @Description Get user by id
// @ID get-string-by-int
// @Accept json
// @Produces json
// @Success 200 {object} UserDummy
// @Router /user/{id} [get]
func GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	c.JSON(http.StatusOK, UserDummy{
		ID:   id,
		Name: "Hickar",
	})
}
