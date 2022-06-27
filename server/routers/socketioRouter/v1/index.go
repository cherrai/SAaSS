package socketioRouter

import (
	socketIoControllersV1 "github.com/cherrai/SAaSS/controllers/socketio/v1"
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
)

type V1 struct {
	Server *socketiomid.SocketIoServer
	Router RouterV1
}

type RouterV1 struct {
	Chat string
}

func (v V1) Init() {

	// s.OnConnect(r.Chat, func(s socketio.Conn) error {
	// 	fmt.Println(r.Chat+"连接成功：", s.ID())
	// 	return nil
	// })

	// r := v.Router
	C := v.Server.OnConnect
	C("/", socketIoControllersV1.NewConnect)

}
