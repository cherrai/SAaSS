package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitCloudServiceFile() {
	c := new(controllersV1.CloudServiceFileController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	// r.Engine.GET("/csf/file/get", c.GetFile)

	r.Group.POST(
		role.SetRole("/csf/file/upload", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		c.UploadFile)

	r.Group.GET(
		role.SetRole("/csf/file/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		c.GetFile)

	r.Group.POST(
		role.SetRole("/csf/file/delete", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		c.UploadFile)

	r.Group.POST(
		role.SetRole("/csf/fileInfos/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		c.UploadFile)

}
