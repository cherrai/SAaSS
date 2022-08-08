package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitFile() {
	fc := new(controllersV1.FileController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	r.Group.POST(
		role.SetRole("/file/delete/file", &middleware.RoleOptionsType{
			CheckApp:           true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.DeleteFile)
	r.Group.GET(
		role.SetRole("/file/get/urls", &middleware.RoleOptionsType{
			CheckApp:           true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetUrls)
	r.Group.GET(
		role.SetRole("/file/get/folder/files", &middleware.RoleOptionsType{
			CheckApp:           true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFolderFiles)

}
