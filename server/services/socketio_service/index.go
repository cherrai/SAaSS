package socketio_service

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/routers/socketioRouter"
	"github.com/cherrai/SAaSS/services/gin_service"
	"github.com/cherrai/SAaSS/services/methods"

	socketioMiddleware "github.com/cherrai/SAaSS/services/middleware/socket.io"
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"

	"github.com/cherrai/nyanyago-utils/nlog"
	sso "github.com/cherrai/saki-sso-go"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

var (
	log = nlog.New()
)
var Server *socketio.Server

// var Router *gin.Engine
// 个人也是一个room，roomId：U+UID
// 群组也是一个room，roomId：G+UID
// ChatMessage 发送消息
// 直接发给对应的roomId即可

// InputStatus 发送正在输入状态
// 直接发给对应的roomId即可

// OnlineStatus 在线状态
// 发送给好友关系存在的roomId

func Init() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	gin.SetMode(conf.Config.Server.Mode)

	gin_service.Router = gin.New()

	Server = socketio.NewServer(nil)
	socketIoServer := socketiomid.New(Server)
	// fmt.Println("Server", Server)
	conf.SocketIoServer = socketIoServer
	// 处理中间件
	socketIoServer.Use(socketioMiddleware.RoleMiddleware())
	socketIoServer.Use(socketioMiddleware.Response())
	socketIoServer.Use(socketioMiddleware.Error())
	socketIoServer.Use(socketioMiddleware.Decryption())

	gin_service.SocketIoServer = Server
	// // redis 适配器
	_, err := Server.Adapter(&socketio.RedisAdapterOptions{
		Addr:    conf.Config.Redis.Addr,
		Prefix:  "socket.io",
		Network: "tcp",
	})

	// fmt.Println("redis Adapter:", ok)

	if err != nil {
		log.Error("error:", err)
	}
	socketioRouter.InitRouter()

	// // 连接成功
	// Server.OnConnect("/", func(s socketio.Conn) error {

	// })

	// 接收”bye“事件
	Server.OnEvent("/", "bye", func(s socketio.Conn, msg string) string {
		last := s.Context().(string)
		s.Emit("bye", msg)

		log.Info("============>", last, msg)
		//s.Close()
		return last
	})

	// 连接错误
	Server.OnError("/", func(s socketio.Conn, e error) {
		c := socketiomid.ConnContext{
			ServerContext: socketIoServer,
			Conn:          s,
		}
		getUserInfo := c.GetSessionCache("userInfo")
		if getUserInfo != nil {
			userInfo := getUserInfo.(*sso.UserInfo)
			c.ClearAllCustomId(methods.GetUserRoomId(userInfo.Uid))
		}
		c.ClearSessionCache()
		log.Error("连接错误:", e)
	})
	// 关闭连接
	Server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		c := socketiomid.ConnContext{
			ServerContext: socketIoServer,
			Conn:          s,
		}
		getUserInfo := c.GetSessionCache("userInfo")
		if getUserInfo != nil {
			userInfo := getUserInfo.(*sso.UserInfo)
			c.ClearAllCustomId(methods.GetUserRoomId(userInfo.Uid))
		}
		c.ClearSessionCache()
		log.Warn(s.ID()+"关闭了连接：", reason)
	})
	go Server.Serve()
	go func() {
		if err := Server.Serve(); err != nil {
			log.Info("[Socket.IO]socketio listen error: %s\n", err)
		}
	}()
	defer Server.Close()

	// http.Handle("/socket.io/", Server)
	// http.Handle("/", http.FileServer(http.Dir("./asset")))
	// log.Println("Serving at localhost:8000...")
	// log.Fatal(http.ListenAndServe(":8000", nil))

	log.Info("[Socket.IO] server created successfully.")

	gin_service.Init()
	// http://localhost:23161
}

func ErrorMiddleware() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("=========Socket.IO ErrorMiddleware=========")
			log.Error(err)
			log.Error("=========Socket.IO ErrorMiddleware=========")
		}
	}()
}
