package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitDownload() {
	c := new(controllersV1.DownloadController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	r.Engine.GET("/s/*any", c.Download)
	r.Engine.GET("/share.html", c.GetShareFilesHtml)

	r.Group.GET(role.SetRole("/share", &middleware.RoleOptionsType{
		CheckApp:           false,
		CheckAppToken:      false,
		Authorize:          false,
		RequestEncryption:  false,
		ResponseEncryption: false,
		ResponseDataType:   "json",
	}), c.GetShareFiles)

	// r.Group.GET(role.SetRole("/share.html", &middleware.RoleOptionsType{
	// 	CheckApp:           false,
	// 	CheckAppToken:      false,
	// 	Authorize:          false,
	// 	RequestEncryption:  false,
	// 	ResponseEncryption: false,
	// 	ResponseDataType:   "json",
	// }), c.GetShareFilesHtml)

}
