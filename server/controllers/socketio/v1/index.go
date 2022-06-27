package socketIoControllersV1

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/methods"
	"strconv"

	// "github.com/cherrai/saki-sso-go"
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"

	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/pasztorpisti/qs"
)

var (
	log = nlog.New()
)

func NewConnect(c *socketiomid.ConnContext) error {
	Conn := c.Conn
	log.Info("正在进行连接.")

	var res response.ResponseProtobufType

	// fmt.Println(msgRes)
	defer func() {
		// fmt.Println("Error middleware.2222222222222")
		if err := recover(); err != nil {
			res.Code = 10001
			res.Data = err.(error).Error()
			Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
			defer Conn.Close()
		}
	}()
	sc := methods.SocketConn{
		Conn: Conn,
	}
	// s.SetContext("dsdsdsd")
	// 申请一个房间

	query := new(typings.SocketEncryptionQuery)

	err := qs.Unmarshal(query, Conn.URL().RawQuery)

	if err != nil {
		res.Code = 10002
		res.Data = err.Error()
		Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
		defer Conn.Close()
		return err
	}
	sc.Query = query
	// log.Info("query", query)

	queryData := new(typings.SocketQuery)
	deQueryDataErr := sc.Decryption(queryData)
	// log.Info("deQueryDataErr", deQueryDataErr != nil, deQueryDataErr)
	if deQueryDataErr != nil {
		res.Code = 10009
		res.Data = deQueryDataErr.Error()
		Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
		defer Conn.Close()
		return deQueryDataErr
	}
	// log.Info("queryData", queryData)
	// res.Code = 10009
	// Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
	// defer Conn.Close()
	// fmt.Println("queryData", queryData)

	// res.Code = 10004
	// res.Data = "SSO Error: " + err.Error()
	// Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
	// defer Conn.Close()

	getUser, err := conf.SSO.Verify(queryData.Token, queryData.DeviceId, queryData.UserAgent)
	if err != nil || getUser == nil || getUser.Payload.Uid == 0 {
		res.Code = 10004
		res.Data = "SSO Error: " + err.Error()
		Conn.Emit(conf.SocketRouterEventNames["Error"], res.GetReponse())
		defer Conn.Close()
		return err
	} else {
		log.Info("UID " + strconv.FormatInt(getUser.Payload.Uid, 10) + ", Connection to Successful.")
		// cc
		// sc.SetUserCache(&ret.Payload)
		log.Info("/ UID", getUser.Payload.Uid)
		c.SetCustomId(methods.GetUserRoomId(getUser.Payload.Uid))
		// c.JoinRoom("/chat", methods.GetUserRoomId(ret.Payload.Uid))
		// fmt.Println("当前房间数：", c.ServerContext.Server.Rooms("/chat"))
		// c.LeaveRoom("/chat", methods.GetUserRoomId(ret.Payload.Uid))
		// fmt.Println("当前房间数：", len(c.ServerContext.Server.Rooms("/chat")))
		c.SetSessionCache("userInfo", &getUser.Payload)
		c.SetSessionCache("deviceId", queryData.DeviceId)
		c.SetSessionCache("userAgent", &queryData.UserAgent)

		log.Info("SocketIO Client连接成功：", Conn.ID())
		// sc.SetUserInfo(&ret.Payload)
	}

	return nil
}
