package routerV1

import (
	controllersV1 "github.com/cherrai/SAaSS/controllers/v1"
	"github.com/cherrai/SAaSS/services/middleware"
)

func (r Routerv1) InitApp() {
	fc := new(controllersV1.AppController)

	role := middleware.RoleMiddlewareOptions{
		BaseUrl: r.BaseUrl,
	}

	r.Group.POST(
		role.SetRole("/app/token/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      false,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetAppToken)

	r.Group.POST(
		role.SetRole("/app/userToken/get", &middleware.RoleOptionsType{
			CheckApp:           true,
			CheckAppToken:      true,
			Authorize:          false,
			RequestEncryption:  false,
			ResponseEncryption: false,
			ResponseDataType:   "json",
		}),
		fc.GetUserToken)

}
