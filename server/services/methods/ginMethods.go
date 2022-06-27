package methods

import (
	"github.com/cherrai/saki-sso-go"
	"github.com/gin-gonic/gin"
)

type Gin struct {
	Context *gin.Context
}

func (g *Gin) GetUserInfo() sso.UserInfo {
	getUserInfo, _ := g.Context.Get("userInfo")

	return getUserInfo.(sso.UserInfo)
}
