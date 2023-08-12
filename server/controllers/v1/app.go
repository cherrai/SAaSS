package controllersV1

import (
	"path"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid"
)

type AppController struct {
}

func (dc *AppController) GetAppToken(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := typings.AppTokenInfo{
		AppId:    c.GetString("appId"),
		AppKey:   c.GetString("appKey"),
		RootPath: c.PostForm("rootPath"),
		UserId:   c.PostForm("userId"),
	}
	log.Info(data)

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.AppKey, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	log.Info(data.AppId, data.AppKey)

	rKey := conf.Redisdb.GetKey("AppToken")
	data.RootPath = path.Join("/", data.RootPath)
	ck := shortuuid.New()
	// log.Info(rKey, data.AppId+data.Uid)
	// fc20c5cc-c567-50e8-90fc-5a0eb1ed6316100000

	if err = conf.Redisdb.SetStruct(rKey.GetKey(ck), &data, rKey.GetExpiration()); err != nil {
		res.Errors(err)
		res.Code = 10001
		res.Call(c)
		return
	}
	res.Data = map[string]interface{}{
		"token":    ck,
		"deadline": time.Now().Unix() + 5*60,
	}
	res.Code = 200
	res.Call(c)
}

func (dc *AppController) GetUserToken(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := typings.AppTokenInfo{
		AppId:  c.GetString("appId"),
		AppKey: c.GetString("appKey"),
		UserId: c.GetString("userId"),
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.AppKey, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	log.Info(data.AppId, data.AppKey)

	rKey := conf.Redisdb.GetKey("UserToken")
	data.RootPath = path.Join("/", data.RootPath)
	ck := shortuuid.New()
	// log.Info(rKey, data.AppId+data.Uid)
	// fc20c5cc-c567-50e8-90fc-5a0eb1ed6316100000

	if err = conf.Redisdb.Set(rKey.GetKey(ck), data.UserId, rKey.GetExpiration()); err != nil {
		res.Errors(err)
		res.Code = 10001
		res.Call(c)
		return
	}
	res.Data = map[string]interface{}{
		"token":    ck,
		"deadline": time.Now().Unix() + 5*60,
	}
	res.Code = 200
	res.Call(c)
}
