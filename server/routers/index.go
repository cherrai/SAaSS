package routers

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	routerV1 "github.com/cherrai/SAaSS/routers/v1"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	download := new(controllersV1.DownloadController)

	r.GET("/s/*any", download.Download)
	rv1 := routerV1.Routerv1{
		Engine:  r,
		BaseUrl: "/api/v1",
	}
	rv1.Init()
}
