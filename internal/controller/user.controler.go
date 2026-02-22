package controller

import (
	"github.com/gin-gonic/gin"
	bh "golang-base/internal/base"
)

type UserController struct {
	bh.BaseController[UserController]
}

func NewUserController() *UserController {
	return &UserController{}
}

func (h *UserController) Paginate(c *gin.Context) {
}
