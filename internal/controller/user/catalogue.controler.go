package user

import (
	c "golang-base/internal/base"

	"github.com/gin-gonic/gin"
)

type UserCatalogueController struct {
	c.BaseController[UserCatalogueController]
}

func NewUserCatalogueController() *UserCatalogueController {
	return &UserCatalogueController{}
}

func (h *UserCatalogueController) Paginate(c *gin.Context) {
}
