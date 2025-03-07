package response

import (
	"time"

	"github.com/gin-gonic/gin"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

// var (
// 	log = nlog.New()
// )

type ResponseType struct {
	// Code 200, 10004
	Code        int64  `json:"code,omitempty"`
	Msg         string `json:"msg,omitempty"`
	CnMsg       string `json:"cnMsg,omitempty"`
	Error       string `json:"error,omitempty"`
	RequestId   string `json:"requestId,omitempty"`
	RequestTime int64  `json:"requestTime,omitempty"`
	Author      string `json:"author,omitempty"`
	Platform    string `json:"platform,omitempty"`
	// RequestTime int64                  `json:"requestTime"`
	// Author      string                 `json:"author"`
	Data interface{} `json:"data,omitempty"`
}

type H map[string]interface{}
type Any *anypb.Any

func (res *ResponseType) Errors(err error) {
	if err != nil {
		res.Error = err.Error()
	}
}
func (res *ResponseType) Call(c *gin.Context) {

	// Log.Info("setResponse", res.GetResponse())
	c.Set("body", res.GetResponse())
	// fmt.Println("setResponse")
	// c.JSON(http.StatusOK, res.GetResponse())
}

func (res *ResponseType) GetResponse() *ResponseType {
	msg := res.Msg
	cnMsg := res.CnMsg
	if res.Msg == "" {
		res.Msg = "Request success."
	}
	if res.CnMsg == "" {
		res.CnMsg = "请求成功"
	}
	if res.Platform == "" {
		res.Platform = "SAaSS"
	}
	if res.Author == "" {
		res.Author = "Shiina Aiiko."
	}
	res.RequestTime = time.Now().Unix()

	switch res.Code {
	case 200:
	case 10025:
		res.Msg = "Delete failed."
		res.CnMsg = "删除失败."

	case 10024:
		res.Msg = "Content does not exist."
		res.CnMsg = "内容不存在."
	case 10023:
		res.Msg = "password required."
		res.CnMsg = "必须输入密码."
	case 10022:
		res.Msg = "incorrect password."
		res.CnMsg = "密码错误."

	case 10021:
		res.Msg = "Copy or move failed."
		res.CnMsg = "复制或移动失败."

	case 10020:
		res.Msg = "Create folder failed."
		res.CnMsg = "创建文件夹失败."

	case 10019:
		res.Msg = "Create file token failed."
		res.CnMsg = "创建文件Token失败"

	case 10018:
		res.Msg = "Chunksize is inconsistent."
		res.CnMsg = "块大小不一致"

	case 10017:
		res.Msg = "The hash value is inconsistent."
		res.CnMsg = "Hash值不一致"

	case 10016:
		res.Msg = "File upload error."
		res.CnMsg = "文件上传错误"

	case 10015:
		res.Msg = "Failed to verify token."
		res.CnMsg = "Token校验失败"

	case 10014:
		res.Msg = "App does not exist."
		res.CnMsg = "应用不存在"

	case 10013:
		res.Msg = "Route does not exist."
		res.CnMsg = "路由不存在"

	case 10012:
		res.Msg = "Already executed."
		res.CnMsg = "已执行过了"

	case 10011:
		res.Msg = "Update failed."
		res.CnMsg = "更新失败"

	case 10010:
		res.Msg = "Insufficient Privilege."
		res.CnMsg = "权限不足."

	case 10009:
		res.Msg = "Decryption failed."
		res.CnMsg = "解密失败."

	case 10008:
		res.Msg = "Encryption key error."
		res.CnMsg = "秘钥错误."

	case 10007:
		res.Msg = "Encryption key generation failed."
		res.CnMsg = "加密秘钥生成失败"

	case 10006:
		res.Msg = "No more."
		res.CnMsg = "没有更多内容了"

	case 10005:
		res.Msg = "Repeat request."
		res.CnMsg = "重复请求"

	case 10004:
		res.Msg = "Login error."
		res.CnMsg = "登陆信息错误"

	case 10001:
		res.Msg = "Request error."
		res.CnMsg = "请求失败"

	case 10002:
		res.Msg = "Parameter error."
		res.CnMsg = "参数错误"

	default:
		res.Msg = "Request error."
		res.CnMsg = "请求失败"

	}
	if res.Code == 0 {
		res.Code = 10001
	}

	if msg != "" {
		res.Msg = msg
	}
	if cnMsg != "" {
		res.CnMsg = cnMsg
	}

	return res
}
