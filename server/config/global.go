package conf

import (
	"github.com/cherrai/SAaSS/services/encryption"
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
)

var SocketIoServer *socketiomid.SocketIoServer

var EncryptionClient *encryption.EncryptionOption
