package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitFolder() {
	fc := new(controllersV1.FolderController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	r.Group.POST(
		role.SetRole("/folder/new", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.NewFolder)

	r.Group.POST(
		role.SetRole("/folder/rename", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.RenameFolder)

	r.Group.POST(
		role.SetRole("/folder/moveToTrash", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.MoveFoldersToTrash)

	// r.Group.POST(
	// 	role.SetRole("/folder/rootPathToken/get", &middleware.RoleOptionsType{
	// 		CheckApp:           true,
	// 		CheckAppToken:      true,
	// 		Authorize:          false,
	// 		RequestEncryption:  false,
	// 		ResponseEncryption: false,
	// 		ResponseDataType:   "json",
	// 	}),
	// 	fc.GetRootFolderToken)

	r.Group.POST(
		role.SetRole("/folder/restore", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.RestoreFolder)

	r.Group.POST(
		role.SetRole("/folder/delete", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.DeleteFolders)

	r.Group.POST(
		role.SetRole("/folder/share/set", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.SetFolderSharing)

	r.Group.POST(
		role.SetRole("/folder/password/set", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.SetFolderPassword)

	r.Group.POST(
		role.SetRole("/folder/copy", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.CopyFolder)

	r.Group.POST(
		role.SetRole("/folder/move", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.MoveFolder)

	r.Group.GET(
		role.SetRole("/folder/list/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GerFolderList)

	r.Group.GET(
		role.SetRole("/folder/shortid/get", &middleware.RoleOptionsType{
			CheckApp:           false,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFolderByShortId)

	r.Group.GET(
		role.SetRole("/folder/list/shortid/get", &middleware.RoleOptionsType{
			CheckApp:           false,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetFolderListWithShortId)

	r.Group.GET(
		role.SetRole("/folder/recyclebin/list/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetRecyclebinFolderList)

}
