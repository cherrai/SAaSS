package socketioRouter

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/routers/socketioRouter/v1"
)

func InitRouter() {
	// fmt.Println(conf.SocketIoServer)
	rv1 := socketioRouter.V1{
		Server: conf.SocketIoServer,
		Router: socketioRouter.RouterV1{
			Chat: "/chat",
		},
	}
	rv1.Init()
}
