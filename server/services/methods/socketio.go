package methods

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/encryption"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"

	socketio "github.com/googollee/go-socket.io"
)

type SocketConn struct {
	Conn      socketio.Conn
	EventName string
	Data      map[string]interface{}
	Query     *typings.SocketEncryptionQuery
}

func (s *SocketConn) Emit(data response.ResponseType) {
	var res response.ResponseType = response.ResponseType{
		Code:      data.Code,
		Data:      data.Data,
		RequestId: s.Data["requestId"].(string),
	}

	s.Conn.Emit(s.EventName, res.GetResponse())
}

func GetCallUserRoomId(userIds []int64) string {
	// fmt.Println(userIds)
	sort.SliceStable(userIds, func(i, j int) bool {
		return userIds[i] < userIds[j]
	})
	userIdsStr := FormatInt64ArrToString(userIds, "")
	// fmt.Println("userIdsStr2", userIdsStr)
	h := md5.New()
	// io.WriteString(h, "The fog is getting thicker!")
	io.WriteString(h, "user")
	io.WriteString(h, userIdsStr)
	roomId := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	return roomId
}
func GetUserRoomId(uid int64) string {
	// fmt.Println("userIdsStr2", userIdsStr)
	h := md5.New()
	// io.WriteString(h, "The fog is getting thicker!")
	io.WriteString(h, "user")
	io.WriteString(h, strconv.FormatInt(uid, 10))
	roomId := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	return roomId
}

func MessageToSocket(namespace string, eventName string, msg interface{}) {
	fmt.Println(namespace, eventName)
	// isPush := conf.SocketIoServer.Server.BroadcastToRoom(namespace, roomId, eventName, msg)
	// fmt.Println("isPush", isPush)
}

func (s *SocketConn) Decryption(data *typings.SocketQuery) error {
	getUserAesKey, getUserAesKeyErr := conf.EncryptionClient.GetUserAesKeyWithAesKey(s.Query.Key)
	if getUserAesKeyErr != nil {
		return errors.New("Failed to get user aes key: " + getUserAesKeyErr.Error())
	}
	aes := encryption.AesEncrypt{
		Key:  getUserAesKey,
		Mode: "CFB",
	}
	deStr, deStrErr := aes.DecryptWithString(s.Query.Data)
	if deStrErr != nil {
		fmt.Println("deStrErr", deStrErr)
		return errors.New("Decryption failed: " + deStrErr.Error())
	}
	err := json.Unmarshal([]byte(deStr), data)
	if err != nil {
		return err
	}
	return nil
}
