package socketiomid

import (
	"encoding/base64"
	"sync"

	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/cherrai/nyanyago-utils/nmap"
	socketio "github.com/googollee/go-socket.io"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	log = nlog.New()
)

type SocketIoServer struct {
	Server             *socketio.Server
	handlers           middlewareFuncType
	disconnectHandlers map[string]([](DisconnectHandlerFunc))
	connectHandlers    map[string]([](ConnectHandlerFunc))
	rooms              map[string](map[string](map[string]*ConnContext))

	sockets      socketsType
	SessionCache SocketCacheType
}
type SocketCacheItemType map[string]interface{}
type SocketCacheType map[string]SocketCacheItemType

type DisconnectHandlerFunc func(s *ConnContext, reason string)
type ConnectHandlerFunc func(s *ConnContext) error
type HandlerFunc func(s *ConnContext) error

type middlewareFuncType []HandlerFunc
type socketsType map[string]socketsNamespaceType

// nammespace+customId+deviceId+deviceType
type socketsNamespaceType map[string]*ConnContext

func New(server *socketio.Server) *SocketIoServer {
	s := new(SocketIoServer)
	s.Server = server
	s.disconnectHandlers = map[string]([](DisconnectHandlerFunc)){}
	s.connectHandlers = map[string]([](ConnectHandlerFunc)){}
	s.rooms = map[string](map[string](map[string]*ConnContext)){}
	s.init()

	return s
}
func (s *SocketIoServer) init() {

}

func (s *SocketIoServer) Router(namespace string, eventName string, f HandlerFunc) {
	s.Server.OnEvent(namespace, eventName, func(sc socketio.Conn, params interface{}) error {
		// fmt.Println("Router函数")
		initData := make(map[string]interface{})
		var m sync.RWMutex
		conn := ConnContext{
			ServerContext: s,
			Conn:          sc,
			index:         -1,
			namespace:     namespace,
			eventName:     eventName,
			params:        params.(map[string]interface{}),
			data:          initData,
			f:             f,
			mu:            m,
			runMiddleware: true,
		}
		conn.Next()
		return conn.err
	})
}
func (s *SocketIoServer) OnConnect(namespace string, f func(c *ConnContext) error) {
	if s.connectHandlers[namespace] == nil {
		s.connectHandlers[namespace] = [](ConnectHandlerFunc){}
	}
	s.connectHandlers[namespace] = append(s.connectHandlers[namespace], f)

	s.Server.OnConnect(namespace, func(sc socketio.Conn) error {
		var m sync.RWMutex
		conn := ConnContext{
			ServerContext: s,
			Conn:          sc,
			index:         -1,
			namespace:     namespace,
			data:          make(map[string]interface{}),
			mu:            m,
			runMiddleware: false,
		}

		for _, f := range s.connectHandlers[namespace] {
			f(&conn)
		}
		return conn.err
	})
}

func (s *SocketIoServer) OnDisconnect(namespace string, f func(c *ConnContext, reason string)) {
	if s.disconnectHandlers[namespace] == nil {
		s.disconnectHandlers[namespace] = [](DisconnectHandlerFunc){}
	}
	s.disconnectHandlers[namespace] = append(s.disconnectHandlers[namespace], f)

	s.Server.OnDisconnect(namespace, func(sc socketio.Conn, reason string) {

		var m sync.RWMutex
		conn := ConnContext{
			ServerContext: s,
			Conn:          sc,
			index:         -1,
			namespace:     namespace,
			data:          make(map[string]interface{}),
			mu:            m,
			runMiddleware: false,
		}
		// 删除该实例的所有Room
		for _, f := range s.disconnectHandlers[namespace] {
			f(&conn, reason)
		}

		conn.LeaveAllRoom()

	})
}

func (s *SocketIoServer) Use(f HandlerFunc) {
	s.handlers = append(s.handlers, f)
}

type ConnContext struct {
	ServerContext *SocketIoServer
	Conn          socketio.Conn
	index         int
	namespace     string
	eventName     string
	customId      string
	params        map[string]interface{}
	data          map[string]interface{}
	f             HandlerFunc
	mu            sync.RWMutex
	runMiddleware bool
	err           error
}

func (c *ConnContext) Next() {
	if !c.runMiddleware {
		c.err = c.f(c)
		return
	}
	c.index++
	if c.index < len(c.ServerContext.handlers) {
		c.err = c.ServerContext.handlers[c.index](c)
	} else {
		c.err = c.f(c)
	}
}

// map[ID]()

// 连接初始化后所存储的缓存，断开连接后无效
// 期间存储的数据在整个连接期间均有效
// 后面均缓存到Redis
func (c *ConnContext) SetSessionCache(key string, data interface{}) {
	// fmt.Println("SocketUserCache", c.ServerContext, c.ServerContext.SessionCache)
	if c.ServerContext.SessionCache == nil {
		c.ServerContext.SessionCache = make(SocketCacheType)
	}
	if c.ServerContext.SessionCache[c.Conn.ID()] == nil {
		c.ServerContext.SessionCache[c.Conn.ID()] = make(SocketCacheItemType)
	}
	c.ServerContext.SessionCache[c.Conn.ID()][key] = data
}

func (c *ConnContext) GetSessionCacheWithConnId(connId string, key string) interface{} {
	if c.ServerContext.SessionCache == nil || c.Conn == nil {
		return nil
	}
	_, ok := c.ServerContext.SessionCache[connId]
	if !ok {
		return nil
	}
	return c.ServerContext.SessionCache[connId][key]
}

func (c *ConnContext) GetSessionCache(key string) interface{} {
	if c.ServerContext.SessionCache == nil || c.Conn == nil {
		return nil
	}
	// Log.Info(" c.ServerContext.SessionCache", c.ServerContext.SessionCache)
	// Log.Info("c.Conn", c.Conn)
	// Log.Info("c.Conn.ID()", c.Conn.ID())
	_, ok := c.ServerContext.SessionCache[c.Conn.ID()]
	if !ok {
		return nil
	}
	return c.ServerContext.SessionCache[c.Conn.ID()][key]
}

// 通过监听disconnect的方式解决
func (c *ConnContext) ClearSessionCache() bool {

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("=========ClearSessionCache Error=========")
	// 		fmt.Println(c)
	// 		fmt.Println(c.ServerContext)
	// 		fmt.Println(c.Conn)
	// 		fmt.Println(err)
	// 		fmt.Println("=========ClearSessionCache Error=========")
	// 	}
	// }()
	if c.Conn == nil || c.ServerContext == nil || c.ServerContext.SessionCache == nil {
		return false
	}
	log.Info("c.ServerContext.SessionCache", c.Conn.ID(), c.ServerContext.SessionCache)
	_, ok := c.ServerContext.SessionCache[c.Conn.ID()]
	if !ok {
		return false
	}
	delete(c.ServerContext.SessionCache, c.Conn.ID())
	return true
}

// 设置自定义ID
func (c *ConnContext) SetCustomId(id string) {
	if c.ServerContext.sockets == nil {
		c.ServerContext.sockets = make(socketsType)
	}
	if c.ServerContext.sockets[c.namespace] == nil {
		c.ServerContext.sockets[c.namespace] = make(socketsNamespaceType)
	}
	c.customId = id
	c.ServerContext.sockets[c.namespace][id] = c

	// 如果socket没了，要删除，未来
}

// 删除所有命名空间的自定义ID
func (c *ConnContext) ClearAllCustomId(id string) bool {
	if c.Conn == nil || c.ServerContext == nil || c.ServerContext.sockets == nil {
		return false
	}
	for k, v := range c.ServerContext.sockets {
		nsMap := v
		if nsMap == nil || nsMap[id] == nil || nsMap[id].Conn == nil {
			return false
		}
		_, ok := nsMap[id]
		if !ok {
			return false
		}
		delete(c.ServerContext.sockets[k], id)
	}
	return true
}

// func (c *ConnContext) GetCustomId(conn socketio.Conn) string {
// 	// fmt.Println(customId, eventName, msg)
// 	nsMap := c.ServerContext.sockets[c.namespace]
// 	// fmt.Println("nsMap", nsMap)
// 	// log.Info("nsMap[customId]", c.namespace, nsMap)
// 	if nsMap == nil || nsMap[customId] == nil || nsMap[customId].Conn == nil {
// 		return false
// 	}
// 	nsMap[customId].Conn.Emit(eventName, msg)
// 	return true
// }

// 发送信息前检测是否存在
func (c *ConnContext) NamespaceToEmit(customId string, eventName string, msg interface{}) bool {
	// fmt.Println(customId, eventName, msg)
	nsMap := c.ServerContext.sockets[c.namespace]
	// fmt.Println("nsMap", nsMap)
	// log.Info("nsMap[customId]", c.namespace, nsMap)
	if nsMap == nil || nsMap[customId] == nil || nsMap[customId].Conn == nil {
		return false
	}
	nsMap[customId].Conn.Emit(eventName, msg)
	return true
}
func (c *ConnContext) Emit(namespace string, customId string, eventName string, msg interface{}) bool {
	nsMap := c.ServerContext.sockets[namespace]
	// log.Info("nsMap", c.ServerContext.sockets, nsMap)
	if nsMap == nil || nsMap[customId] == nil || nsMap[customId].Conn == nil {
		return false
	}
	nsMap[customId].Conn.Emit(eventName, msg)
	return true
}

// 获取房间的所有实例
func (c *ConnContext) GetAllConnOfRoomWithNamespace(roomId string) []socketio.Conn {
	return c.GetAllConnOfRoom(c.Namespace(), roomId)

}
func (c *ConnContext) GetAllConnOfRoom(namespace string, roomId string) []socketio.Conn {
	conns := []socketio.Conn{}
	c.ServerContext.Server.ForEach(namespace, roomId, func(conn socketio.Conn) {
		conns = append(conns, conn)
	})
	return conns
}

// 获取房间的所有实例个数
func (c *ConnContext) GetRoomLenWithNamespace(roomId string) int {
	return c.GetRoomLen(c.Namespace(), roomId)

}
func (c *ConnContext) GetRoomLen(namespace string, roomId string) int {
	return c.ServerContext.Server.RoomLen(namespace, roomId)
}

// 获取该用户加入的所有房间
func (c *ConnContext) GetRoomsWithNamespace() []string {
	return c.GetRooms(c.Namespace())
}
func (c *ConnContext) GetRooms(namespace string) []string {
	rooms := []string{}
	for mapRoomIdKey, mapRoomId := range c.ServerContext.rooms[namespace] {
		for _, v := range mapRoomId {
			if v.Conn.ID() == c.Conn.ID() {
				rooms = append(rooms, mapRoomIdKey)
				break
			}
		}
	}
	return rooms
}

// 获取房间下的所有用户
func (c *ConnContext) GetAllConnContextInRoomWithNamespace(roomId string) []*ConnContext {
	return c.GetAllConnContextInRoom(c.Namespace(), roomId)
}
func (c *ConnContext) GetAllConnContextInRoom(namespace string, roomId string) []*ConnContext {
	conns := []*ConnContext{}
	for _, conn := range c.ServerContext.rooms[namespace][roomId] {
		conns = append(conns, conn)
	}
	return conns
}

// 发送给房间信息
func (c *ConnContext) BroadcastToRoomWithNamespace(roomId string, eventName string, msg interface{}) bool {
	return c.ServerContext.Server.BroadcastToRoom(c.namespace, roomId, eventName, msg)
}
func (c *ConnContext) BroadcastToRoom(namespace string, roomId string, eventName string, msg interface{}) bool {
	return c.ServerContext.Server.BroadcastToRoom(namespace, roomId, eventName, msg)
}
func (c *ConnContext) ForEachBroadcastToRoomWithNamespace(roomId string, eventName string, msg interface{}) bool {
	return c.ServerContext.Server.BroadcastToRoom(c.namespace, roomId, eventName, msg)
}

// 加入房间
func (c *ConnContext) JoinRoomWithNamespace(roomId string) bool {
	return c.JoinRoom(c.namespace, roomId)
}

func (c *ConnContext) JoinRoom(namespace string, roomId string) bool {

	isJoin := c.ServerContext.Server.JoinRoom(namespace, roomId, c.Conn)
	if isJoin {
		c.addRoom(namespace, roomId)
	}
	return isJoin
}

func (c *ConnContext) addRoom(namespace string, roomId string) bool {
	isJoin := c.ServerContext.Server.JoinRoom(namespace, roomId, c.Conn)
	if isJoin {
		if c.ServerContext.rooms[namespace] == nil {
			c.ServerContext.rooms[namespace] = map[string](map[string]*ConnContext){}
		}
		if c.ServerContext.rooms[namespace][roomId] == nil {
			c.ServerContext.rooms[namespace][roomId] = map[string]*ConnContext{}
		}
		c.ServerContext.rooms[namespace][roomId][c.Conn.ID()] = c
	}
	return isJoin
}

// 预留 离开房间
func (c *ConnContext) NamespaceToLeaveRoom(roomId string) bool {
	return c.LeaveRoom(c.namespace, roomId)
}
func (c *ConnContext) LeaveRoom(namespace string, roomId string) bool {

	isLeave := c.ServerContext.Server.LeaveRoom(namespace, roomId, c.Conn)
	if isLeave {
		c.deleteRoom(namespace, roomId)
	}
	return isLeave
}
func (c *ConnContext) deleteRoom(namespace string, roomId string) bool {
	isLeave := c.ServerContext.Server.LeaveRoom(namespace, roomId, c.Conn)
	if isLeave {
		delete(c.ServerContext.rooms[namespace][roomId], c.Conn.ID())
		if nmap.KeysLenth(c.ServerContext.rooms[namespace][roomId]) == 0 {
			delete(c.ServerContext.rooms[namespace], roomId)
		}
		if nmap.KeysLenth(c.ServerContext.rooms[namespace]) == 0 {
			delete(c.ServerContext.rooms, namespace)
		}
	}
	return isLeave
}

// 离开所有房间
func (c *ConnContext) LeaveAllRoom() {
	for _, v := range c.GetRoomsWithNamespace() {
		c.LeaveRoom(c.Namespace(), v)
	}
}

func (c *ConnContext) Namespace() string {
	return c.namespace
}
func (c *ConnContext) EventName() string {
	return c.eventName
}
func (c *ConnContext) Params() map[string]interface{} {
	return c.params
}
func (c *ConnContext) GetParams(key string) (value interface{}, exists bool) {

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("=========GetParams Error=========")
	// 		fmt.Println(c)
	// 		fmt.Println(c.ServerContext)
	// 		fmt.Println(c.params)
	// 		fmt.Println(err)
	// 		fmt.Println("=========GetParams Error=========")
	// 	}
	// }()
	// fmt.Println("c.params", c)
	// var m sync.RWMutex
	// m.RLock()
	if c != nil && c.params != nil {
		value, exists = c.params[key]
		// fmt.Println("exists", exists, key)
		if !exists {
			return
		}
		return
	}
	value = nil
	exists = false
	return
}
func (c *ConnContext) GetParamsString(key string) string {
	if value, exists := c.GetParams(key); exists {
		return value.(string)
	}
	return ""
}

func (c *ConnContext) Set(k string, data interface{}) {
	c.data[k] = data
}
func (c *ConnContext) GetData(k string) interface{} {
	return c.data[k]
}
func (c *ConnContext) Get(k string) (value interface{}, exists bool) {
	// var m sync.RWMutex
	// m.RLock()
	// fmt.Println("c.data", c.data)
	if c == nil || c.data == nil {
		return
	}
	if value, exists = c.data[k]; exists {
		return
	}
	// m.RUnlock()
	return
}
func (c *ConnContext) GetString(k string) string {

	if value, exists := c.Get(k); exists {
		return value.(string)
	}
	return ""
}
func (c *ConnContext) GetInt(k string) int {
	if value, exists := c.Get(k); exists {
		return value.(int)
	}
	return int(0)
}
func (c *ConnContext) GetInt64(k string) int64 {
	if value, exists := c.Get(k); exists {
		return value.(int64)
	}
	return int64(0)
}
func (c *ConnContext) GetFloat64(k string) float64 {
	if value, exists := c.Get(k); exists {
		return value.(float64)
	}
	return float64(0)
}
func (c *ConnContext) GetBool(k string) bool {
	if value, exists := c.Get(k); exists {
		return value.(bool)
	}
	return false
}
func (c *ConnContext) Body(response interface{}) {
	c.Conn.Emit(c.EventName(), response)
}
func (c *ConnContext) Protobuf(response interface{}) {
	getData, getDataErr := proto.Marshal(response.(protoreflect.ProtoMessage))
	if getDataErr != nil {
		panic(getDataErr)
	}
	enData := base64.StdEncoding.EncodeToString(getData)
	c.Conn.Emit(c.EventName(), enData)
}

// func A() MiddlewareHandlerFunc {
// 	return func(s *ConnContext, data interface{}) {
// 		fmt.Println("A中间件01")
// 		// s.Abort()
// 		fmt.Println("A中间件02")
// 	}
// }

// func B() MiddlewareHandlerFunc {
// 	return func(s *ConnContext, data interface{}) {
// 		fmt.Println("B中间件01")
// 		s.Next()
// 		fmt.Println("B中间件02")
// 	}
// }
