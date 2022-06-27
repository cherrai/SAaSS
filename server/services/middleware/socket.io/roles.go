package socketioMiddleware

import (
	socketiomid "github.com/cherrai/SAaSS/services/nyanyago-utils/socketio-mid"
)

// 默认全部都要进行加密，且全部输出protobuf
// 全部都要进行权限校验
// 所以权限暂时弃用
func RoleMiddleware() socketiomid.HandlerFunc {
	return func(c *socketiomid.ConnContext) error {
		c.Next()
		return nil
	}
}
