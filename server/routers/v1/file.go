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

	r.Group.GET(
		role.SetRole("/file/shortid/get", &middleware.RoleOptionsType{
			CheckApp:           false,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFileByShortId)

	r.Group.POST(
		role.SetRole("/file/moveToTrash", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.MoveFilesToTrash)

	r.Group.POST(
		role.SetRole("/file/restore", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.RestoreFile)

	r.Group.POST(
		role.SetRole("/file/checkExists", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.CheckFileExists)

	r.Group.POST(
		role.SetRole("/file/delete", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.DeleteFiles)

	r.Group.POST(
		role.SetRole("/file/rename", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.RenameFile)

	r.Group.POST(
		role.SetRole("/file/share/set", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.SetFileSharing)

	r.Group.POST(
		role.SetRole("/file/password/set", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.SetFilePassword)

	r.Group.POST(
		role.SetRole("/file/passwordToken/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetPasswordToken)

	r.Group.POST(
		role.SetRole("/file/copy", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.CopyFile)

	r.Group.POST(
		role.SetRole("/file/move", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.MoveFile)

	r.Group.GET(
		role.SetRole("/file/get/urls", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetUrls)
	r.Group.GET(
		role.SetRole("/file/list/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFileList)

	r.Group.GET(
		role.SetRole("/file/list/shortid/get", &middleware.RoleOptionsType{
			CheckApp:           false,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFileListWithShortId)

	r.Group.GET(
		role.SetRole("/file/recent/list/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetRecentFiles)
	r.Group.GET(
		role.SetRole("/file/recyclebin/list/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetRecyclebinFiles)

}
