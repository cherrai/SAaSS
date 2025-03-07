package routerV1

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/gin-gonic/gin"
)

var (
	log = conf.Log
)

type Routerv1 struct {
	Engine  *gin.Engine
	Group   *gin.RouterGroup
	BaseUrl string
}

func (r Routerv1) Init() {
	r.Group = r.Engine.Group(r.BaseUrl)
	r.InitUpload()
	r.InitChunkUpload()
	r.InitFile()
	r.InitFolder()
	r.InitApp()
	r.InitDownload()
	r.InitCloudServiceFile()
}
