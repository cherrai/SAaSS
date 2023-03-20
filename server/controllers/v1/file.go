package controllersV1

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/ncredentials"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileController struct {
}

func (dc *FileController) MoveFilesToTrash(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		Path      string
		UserId    string
		FileNames map[string]string
		RootPath  string
	}{
		AppId:     c.GetString("appId"),
		Path:      c.PostForm("path"),
		UserId:    c.GetString("userId"),
		FileNames: c.PostFormMap("fileNames"),
		RootPath:  c.PostForm("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	if err := fileDbx.MoveFilesToTrash(params.AppId, path.Join(params.RootPath, params.Path), fns, params.UserId); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) CheckFileExists(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		Path      string
		UserId    string
		FileNames map[string]string
		RootPath  string
	}{
		AppId:     c.GetString("appId"),
		Path:      c.PostForm("path"),
		UserId:    c.GetString("userId"),
		FileNames: c.PostFormMap("fileNames"),
		RootPath:  c.PostForm("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	parentFolderId, err := folderDbx.GetParentFolderId(params.AppId, path.Join(params.RootPath, params.Path), false, params.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10001
		res.Call(c)
		return
	}
	files, err := fileDbx.GetFileLisByParentFolderIdOrFileNames(params.AppId, parentFolderId, fns)
	if err != nil {
		res.Errors(err)
		res.Code = 10001
		res.Call(c)
		return
	}
	list := []string{}
	for _, v := range files.List {
		list = append(list, v.FileName)
	}
	res.Data = map[string]interface{}{
		"list":  list,
		"total": len(list),
	}
	res.Call(c)
}

func (dc *FileController) RestoreFile(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		Path      string
		UserId    string
		FileNames map[string]string
		RootPath  string
	}{
		AppId:     c.GetString("appId"),
		Path:      c.PostForm("path"),
		UserId:    c.GetString("userId"),
		FileNames: c.PostFormMap("fileNames"),
		RootPath:  c.PostForm("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	if err := fileDbx.Restore(params.AppId, path.Join(params.RootPath, params.Path), fns, params.UserId); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) DeleteFiles(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		Path      string
		UserId    string
		FileNames map[string]string
		RootPath  string
	}{
		AppId:     c.GetString("appId"),
		Path:      c.PostForm("path"),
		UserId:    c.GetString("userId"),
		FileNames: c.PostFormMap("fileNames"),
		RootPath:  c.PostForm("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	if err := fileDbx.DeleteFiles(params.AppId, path.Join(params.RootPath, params.Path), fns, params.UserId, []int64{-1}); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) RenameFile(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId       string
		UserId      string
		Path        string
		OldFileName string
		NewFileName string
		RootPath    string
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		Path:        c.PostForm("path"),
		OldFileName: c.PostForm("oldFileName"),
		NewFileName: c.PostForm("newFileName"),
		RootPath:    c.PostForm("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.OldFileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.NewFileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)

	// 4、操作数据库
	if err := fileDbx.RenameFile(params.AppId, p, params.OldFileName, params.NewFileName, params.UserId); err != nil {
		res.Error = err.Error()
		res.Code = 10011
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) SetFileSharing(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		UserId    string
		RootPath  string
		Path      string
		FileNames map[string]string
		Status    int64
	}{
		AppId:     c.GetString("appId"),
		UserId:    c.GetString("userId"),
		RootPath:  c.PostForm("rootPath"),
		Path:      c.PostForm("path"),
		FileNames: c.PostFormMap("fileNames"),
		Status:    nint.ToInt64(c.PostForm("status")),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.Status, validation.Type("int64"), validation.Enum([]int64{1, -1}), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	p := path.Join(params.RootPath, params.Path)

	// 4、操作数据库
	if err := fileDbx.SetFileSharing(params.AppId, params.UserId, p, fns, params.Status); err != nil {
		res.Error = err.Error()
		res.Code = 10011
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) SetFilePassword(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		UserId   string
		RootPath string
		Path     string
		FileName string
		Password string
	}{
		AppId:    c.GetString("appId"),
		UserId:   c.GetString("userId"),
		RootPath: c.PostForm("rootPath"),
		Path:     c.PostForm("path"),
		FileName: c.PostForm("fileName"),
		Password: c.PostForm("password"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Password, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)

	if params.Password == "noPassword" {
		params.Password = ""
	}

	// 4、操作数据库
	if err := fileDbx.SetFilePassword(params.AppId, params.UserId, p, params.FileName, params.Password); err != nil {
		res.Error = err.Error()
		res.Code = 10011
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) CopyFile(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		UserId    string
		RootPath  string
		Path      string
		FileNames map[string]string
		NewPath   string
	}{
		AppId:     c.GetString("appId"),
		UserId:    c.GetString("userId"),
		RootPath:  c.PostForm("rootPath"),
		Path:      c.PostForm("path"),
		FileNames: c.PostFormMap("fileNames"),
		NewPath:   c.PostForm("newPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.NewPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)
	np := path.Join(params.RootPath, params.NewPath)

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	if err := fileDbx.CopyFile(params.AppId, params.UserId, p, fns, np); err != nil {
		res.Error = err.Error()
		res.Code = 10021
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) MoveFile(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId     string
		UserId    string
		RootPath  string
		Path      string
		FileNames map[string]string
		NewPath   string
	}{
		AppId:     c.GetString("appId"),
		UserId:    c.GetString("userId"),
		RootPath:  c.PostForm("rootPath"),
		Path:      c.PostForm("path"),
		FileNames: c.PostFormMap("fileNames"),
		NewPath:   c.PostForm("newPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileNames, validation.Required()),
		validation.Parameter(&params.NewPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)
	np := path.Join(params.RootPath, params.NewPath)

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FileNames {
		fns = append(fns, v)
	}
	if err := fileDbx.MoveFile(params.AppId, params.UserId, p, fns, np); err != nil {
		res.Error = err.Error()
		res.Code = 10021
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FileController) GetUrls(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		UserId   string
		Path     string
		FileName string
		RootPath string
	}{
		AppId:    c.GetString("appId"),
		UserId:   c.GetString("userId"),
		Path:     c.Query("path"),
		FileName: c.Query("fileName"),
		RootPath: c.Query("rootPath"),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	p := path.Join(params.RootPath, params.Path)

	// 4、操作数据库
	file, err := fileDbx.GetFileWithFileInfo(params.AppId, p, params.FileName, params.UserId)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	shortUrl := "/s/" + file.ShortId
	if file.AvailableRange.AllowShare == -1 {
		// time.Duration(params.Deadline-time.Now().Unix()) * time.Second

		at := methods.GetTemporaryAccessToken(file.ShortId, time.Now().Add(5*60).Unix())
		shortUrl = shortUrl + "?u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
	}
	res.Code = 200
	res.Data = map[string]interface{}{
		"urls": map[string]string{
			"domainUrl": conf.Config.StaticPathDomain,
			"shortUrl":  shortUrl,
			"url":       "/s" + path.Join(params.Path, file.FileName) + "?a=" + conf.AppList[params.AppId].EncryptionId + "&r=" + params.RootPath,
		},
	}
	res.Call(c)
}

func (dc *FileController) GetFileByShortId(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		RootPath string
		UserId   string
		Id       string
		Password string
		Deadline int64
	}{
		AppId:    c.Query("appId"),
		RootPath: c.Query("rootPath"),
		UserId:   c.GetString("userId"),
		Id:       c.Query("id"),
		Password: c.Query("password"),
		Deadline: nint.ToInt64(c.Query("deadline")),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.AppId = t.AppId
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		// validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Id, validation.Type("string"), validation.Required()),
		// validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	// 4、操作数据库
	v, err := fileDbx.GetFileWithShortId(params.Id)
	log.Info("GetFileWithShortId", v, v.AvailableRange.AllowShare, err)
	if err != nil || v == nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}

	if v.AvailableRange.AllowShare == -1 {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}

	if v.AvailableRange.Password != "" && params.Password == "" {
		res.Data = map[string]interface{}{
			"id":       v.Id,
			"shortId":  v.ShortId,
			"fileName": v.FileName,
			"availableRange": map[string]interface{}{
				"authorId": v.AvailableRange.AuthorId,
			},
		}
		res.Code = 10023
		res.Call(c)
		return
	} else {
		if v.AvailableRange.Password != "" && params.Password == v.AvailableRange.Password {
			if err != nil {
				res.Errors(err)
				res.Code = 10022
				res.Call(c)
				return
			}
		}
		user := ""
		passwordToken := ""
		shortUrl := "/s/" + v.ShortId
		url := "/s/" + v.FileName + "?sid=" + v.ShortId
		if v.AvailableRange.Password != "" {
			if err := validation.ValidateStruct(
				&params,
				validation.Parameter(&params.Password, validation.Type("string"), validation.Required()),
				validation.Parameter(&params.Deadline, validation.Type("int64"), validation.Required()),
			); err != nil {
				res.Errors(err)
				res.Code = 10002
				res.Call(c)
				return
			}
			t := time.Duration(params.Deadline-time.Now().Unix()) * time.Second

			appKey := ""
			for _, sv := range conf.AppList {
				if sv.AppId == v.AppId {
					appKey = sv.AppKey
					break
				}
			}
			user, passwordToken = ncredentials.GenerateCredentials(appKey+v.AvailableRange.Password, t)
			if err != nil {
				res.Errors(err)
				res.Code = 10006
				res.Call(c)
				return
			}
			shortUrl = shortUrl + "?u=" + user + "&p=" + passwordToken
			url = url + "&u=" + user + "&p=" + passwordToken
		}
		staticFileList, err := fileDbx.GetStaticFileListWithHash([]string{
			v.Hash,
		})
		if err != nil || len(staticFileList) == 0 {
			res.Errors(err)
			res.Code = 10001
			res.Call(c)
			return
		}
		// log.Info("file", file)
		sv := staticFileList[0]

		res.Data = map[string]interface{}{
			"id":       v.Id,
			"shortId":  v.ShortId,
			"fileName": v.FileName,
			// "path":           params.Path,
			"parentFolderId": v.ParentFolderId.Hex(),
			"availableRange": map[string]interface{}{
				"visitCount":     v.AvailableRange.VisitCount,
				"expirationTime": v.AvailableRange.ExpirationTime,
				"password":       v.AvailableRange.Password,
				"allowShare":     v.AvailableRange.AllowShare,
				"shareUsers":     v.AvailableRange.ShareUsers,
				"authorId":       v.AvailableRange.AuthorId,
			},
			"usage": map[string]interface{}{
				"visitCount": v.Usage.VisitCount,
			},
			"createTime":     v.CreateTime,
			"lastUpdateTime": v.LastUpdateTime,
			"deleteTime":     v.DeleteTime,
			"urls": map[string]string{
				"domainUrl": conf.Config.StaticPathDomain,
				"shortUrl":  shortUrl,
				"url":       url,
			},
			"fileInfo": map[string]interface{}{
				"name":         sv.FileInfo.Name,
				"size":         sv.FileInfo.Size,
				"type":         sv.FileInfo.Type,
				"suffix":       sv.FileInfo.Suffix,
				"lastModified": sv.FileInfo.LastModified,
				"hash":         sv.FileInfo.Hash,
				"width":        sv.FileInfo.Width,
				"height":       sv.FileInfo.Height,
			},
		}
	}

	res.Code = 200
	res.Call(c)
}

func (dc *FileController) GetFileList(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId            string
		Path             string
		RootPath         string
		UserId           string
		ParentFolderPath string
	}{
		AppId:            c.Query("appId"),
		RootPath:         c.Query("rootPath"),
		UserId:           c.GetString("userId"),
		Path:             c.Query("path"),
		ParentFolderPath: "/",
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.AppId = t.AppId
		params.RootPath = t.RootPath
	}

	log.Info("params", params)
	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	params.ParentFolderPath = path.Join(params.RootPath, params.Path)

	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候
	log.Info(params.ParentFolderPath)
	parentFolderId, err := folderDbx.GetParentFolderId(params.AppId, params.ParentFolderPath, false, params.UserId)
	if err != nil || parentFolderId == primitive.NilObjectID {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}
	// 4、操作数据库
	file, err := fileDbx.GetFileLisByParentFolderId(params.AppId, parentFolderId)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	res.Code = 200
	// log.Info("file", file)
	tempList := []map[string]interface{}{}

	if len(file.Total) == 0 {
		res.Data = map[string]interface{}{
			"total": 0,
			"list":  tempList,
		}
		res.Call(c)
		return
	}

	hashList := []string{}
	for _, v := range file.List {
		hashList = append(hashList, v.Hash)
	}

	staticFileList, err := fileDbx.GetStaticFileListWithHash(hashList)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	for _, v := range file.List {
		for _, sv := range staticFileList {
			if sv.FileInfo.Hash == v.Hash {
				// pd := ""
				// if v.AvailableRange.Password != "" {
				// 	pd = v.AvailableRange.Password[0:2] + "******" + v.AvailableRange.Password[len(v.AvailableRange.Password)-3:len(v.AvailableRange.Password)]
				// }

				shortUrl := "/s/" + v.ShortId
				if v.AvailableRange.AllowShare == -1 {
					at := methods.GetTemporaryAccessToken(v.ShortId, time.Now().Add(5*60).Unix())
					shortUrl = shortUrl + "?u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
				}

				tempList = append(tempList, map[string]interface{}{
					"id":             v.Id,
					"shortId":        v.ShortId,
					"fileName":       v.FileName,
					"path":           params.Path,
					"parentFolderId": v.ParentFolderId.Hex(),
					"availableRange": map[string]interface{}{
						"visitCount":     v.AvailableRange.VisitCount,
						"expirationTime": v.AvailableRange.ExpirationTime,
						"password":       v.AvailableRange.Password,
						"allowShare":     v.AvailableRange.AllowShare,
						"shareUsers":     v.AvailableRange.ShareUsers,
						"authorId":       v.AvailableRange.AuthorId,
					},
					"usage": map[string]interface{}{
						"visitCount": v.Usage.VisitCount,
					},
					"createTime":     v.CreateTime,
					"lastUpdateTime": v.LastUpdateTime,
					"deleteTime":     v.DeleteTime,
					"urls": map[string]string{
						"domainUrl": conf.Config.StaticPathDomain,
						"shortUrl":  shortUrl,
						"url":       "/s" + path.Join(params.Path, v.FileName) + "?a=" + conf.AppList[v.AppId].EncryptionId + "&r=" + params.RootPath,
					},
					"fileInfo": map[string]interface{}{
						"name":         sv.FileInfo.Name,
						"size":         sv.FileInfo.Size,
						"type":         sv.FileInfo.Type,
						"suffix":       sv.FileInfo.Suffix,
						"lastModified": sv.FileInfo.LastModified,
						"hash":         sv.FileInfo.Hash,
						"width":        sv.FileInfo.Width,
						"height":       sv.FileInfo.Height,
					},
				})
				break
			}
		}
	}
	res.Data = map[string]interface{}{
		"total": file.Total[0].Count,
		"list":  tempList,
	}
	res.Call(c)
}

func (dc *FileController) GetFileListWithShortId(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	params := struct {
		AppId       string
		AppKey      string
		UserId      string
		Id          string
		AccessToken map[string]string
		Deadline    int64
	}{
		AppId:       c.GetString("appId"),
		AppKey:      c.GetString("appKey"),
		UserId:      c.GetString("userId"),
		Id:          c.Query("id"),
		AccessToken: c.QueryMap("accessToken"),
		Deadline:    nint.ToInt64(c.Query("deadline")),
	}
	at := c.Query("accessToken")
	if at != "" {
		err := json.Unmarshal([]byte(at), &params.AccessToken)
		if err != nil {
			res.Error = err.Error()
			res.Code = 10002
			res.Call(c)
			return
		}
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&params,
		validation.Parameter(&params.Id, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.AccessToken, validation.Required()),
		validation.Parameter(&params.Deadline, validation.Type("int64"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	user := params.AccessToken["user"]
	temporaryAccessToken := params.AccessToken["temporaryAccessToken"]

	if !methods.VerfiyTemporaryAccessToken(params.Id, user, temporaryAccessToken) {
		c.String(http.StatusNotFound, "")
		return
	}

	folder, err := folderDbx.GetFolderWithShortId(params.Id)
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}

	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候

	file, err := fileDbx.GetFileLisByParentFolderId(folder.AppId, folder.Id)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	res.Code = 200
	// log.Info("file", file)
	tempList := []map[string]interface{}{}

	if len(file.Total) == 0 {
		res.Data = map[string]interface{}{
			"total": 0,
			"list":  tempList,
		}
		res.Call(c)
		return
	}

	hashList := []string{}
	for _, v := range file.List {
		hashList = append(hashList, v.Hash)
	}

	staticFileList, err := fileDbx.GetStaticFileListWithHash(hashList)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	for _, v := range file.List {
		for _, sv := range staticFileList {
			if sv.FileInfo.Hash == v.Hash {
				// pd := ""
				// if v.AvailableRange.Password != "" {
				// 	pd = v.AvailableRange.Password[0:2] + "******" + v.AvailableRange.Password[len(v.AvailableRange.Password)-3:len(v.AvailableRange.Password)]
				// }

				// shortUrl := "/s/" + v.ShortId
				// url := "/s/" + v.FileName + "?sid=" + v.ShortId
				at := methods.GetTemporaryAccessToken(v.ShortId, params.Deadline)
				log.Info(params.Deadline)
				shortUrl := "/s/" + v.ShortId + "?u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
				url := "/s/" + v.FileName + "?sid=" + v.ShortId + "&u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
				tempList = append(tempList, map[string]interface{}{
					"id":       v.Id,
					"shortId":  v.ShortId,
					"fileName": v.FileName,
					// "path":           params.Path,
					"parentFolderId": v.ParentFolderId.Hex(),
					"availableRange": map[string]interface{}{
						"visitCount":     v.AvailableRange.VisitCount,
						"expirationTime": v.AvailableRange.ExpirationTime,
						"password":       v.AvailableRange.Password,
						"allowShare":     v.AvailableRange.AllowShare,
						"shareUsers":     v.AvailableRange.ShareUsers,
						"authorId":       v.AvailableRange.AuthorId,
					},
					"usage": map[string]interface{}{
						"visitCount": v.Usage.VisitCount,
					},
					"createTime":     v.CreateTime,
					"lastUpdateTime": v.LastUpdateTime,
					"deleteTime":     v.DeleteTime,
					"urls": map[string]string{
						"domainUrl": conf.Config.StaticPathDomain,
						"shortUrl":  shortUrl,
						"url":       url,
					},
					"fileInfo": map[string]interface{}{
						"name":         sv.FileInfo.Name,
						"size":         sv.FileInfo.Size,
						"type":         sv.FileInfo.Type,
						"suffix":       sv.FileInfo.Suffix,
						"lastModified": sv.FileInfo.LastModified,
						"hash":         sv.FileInfo.Hash,
						"width":        sv.FileInfo.Width,
						"height":       sv.FileInfo.Height,
					},
				})
				break
			}
		}
	}
	res.Data = map[string]interface{}{
		"total": file.Total[0].Count,
		"list":  tempList,
	}
	res.Call(c)
}

func (dc *FileController) GetRecentFiles(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		Path     string
		RootPath string
		UserId   string
		PageNum  int64
		PageSize int64
	}{
		AppId:    c.Query("appId"),
		RootPath: c.Query("rootPath"),
		UserId:   c.GetString("userId"),
		Path:     c.Query("path"),
		PageNum:  nint.ToInt64(c.Query("pageNum")),
		PageSize: nint.ToInt64(c.Query("pageSize")),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.AppId = t.AppId
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.PageNum, validation.Type("int64"),
			validation.GreaterEqual(1), validation.Required()),
		validation.Parameter(&params.PageSize, validation.Type("int64"),
			validation.GreaterEqual(1), validation.LessEqual(50), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)
	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候
	parentFolderId, err := folderDbx.GetParentFolderId(params.AppId, p, false, params.UserId)
	if err != nil || parentFolderId == primitive.NilObjectID {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	getFolders, err := folderDbx.GetFolderTreeByParentFolderId(params.AppId, parentFolderId, []int64{1, 0, -1, -2})
	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	parentFolderIdList := []primitive.ObjectID{
		parentFolderId,
	}

	pathMap := map[primitive.ObjectID]string{
		parentFolderId: params.Path,
	}

	for _, v := range getFolders {
		pathMap[v.Id] = path.Join(pathMap[v.ParentFolderId], v.FolderName)
		parentFolderIdList = append(parentFolderIdList, v.Id)
	}
	log.Info("pathMap", pathMap)
	// 4、操作数据库
	files, err := fileDbx.GetFileLisByParentFolderIdList(
		params.AppId,
		parentFolderIdList,
		params.PageNum,
		params.PageSize,
		[]int64{1, 0})

	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	res.Code = 200
	// log.Info("file", file)
	tempList := []map[string]interface{}{}

	if len(files) == 0 {
		res.Data = map[string]interface{}{
			"total": 0,
			"list":  len(files),
		}
		res.Call(c)
		return
	}

	hashList := []string{}
	for _, v := range files {
		hashList = append(hashList, v.Hash)
	}

	staticFileList, err := fileDbx.GetStaticFileListWithHash(hashList)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	for _, v := range files {
		for _, sv := range staticFileList {
			if sv.FileInfo.Hash == v.Hash {
				// pd := ""
				// if v.AvailableRange.Password != "" {
				// 	pd = v.AvailableRange.Password[0:2] + "******" + v.AvailableRange.Password[len(v.AvailableRange.Password)-3:len(v.AvailableRange.Password)]
				// }
				tempList = append(tempList, map[string]interface{}{
					"id":             v.Id,
					"shortId":        v.ShortId,
					"fileName":       v.FileName,
					"path":           pathMap[v.ParentFolderId],
					"parentFolderId": v.ParentFolderId.Hex(),
					"availableRange": map[string]interface{}{
						"visitCount":     v.AvailableRange.VisitCount,
						"expirationTime": v.AvailableRange.ExpirationTime,
						"password":       v.AvailableRange.Password,
						"allowShare":     v.AvailableRange.AllowShare,
						"shareUsers":     v.AvailableRange.ShareUsers,
						"authorId":       v.AvailableRange.AuthorId,
					},
					"usage": map[string]interface{}{
						"visitCount": v.Usage.VisitCount,
					},
					"createTime":     v.CreateTime,
					"lastUpdateTime": v.LastUpdateTime,
					"deleteTime":     v.DeleteTime,
					"urls": map[string]string{
						"domainUrl": conf.Config.StaticPathDomain,
						"shortUrl":  "/s/" + v.ShortId,
						"url":       "/s" + path.Join(pathMap[v.ParentFolderId], v.FileName) + "?a=" + conf.AppList[v.AppId].EncryptionId + "&r=" + params.RootPath,
					},
					"fileInfo": map[string]interface{}{
						"name":         sv.FileInfo.Name,
						"size":         sv.FileInfo.Size,
						"type":         sv.FileInfo.Type,
						"suffix":       sv.FileInfo.Suffix,
						"lastModified": sv.FileInfo.LastModified,
						"hash":         sv.FileInfo.Hash,
						"width":        sv.FileInfo.Width,
						"height":       sv.FileInfo.Height,
					},
				})
				break
			}
		}
	}
	res.Data = map[string]interface{}{
		"total": len(files),
		"list":  tempList,
	}
	res.Call(c)
}

func (dc *FileController) GetRecyclebinFiles(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		Path     string
		RootPath string
		UserId   string
		PageNum  int64
		PageSize int64
	}{
		AppId:    c.Query("appId"),
		RootPath: c.Query("rootPath"),
		UserId:   c.GetString("userId"),
		Path:     c.Query("path"),
		PageNum:  nint.ToInt64(c.Query("pageNum")),
		PageSize: nint.ToInt64(c.Query("pageSize")),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.AppId = t.AppId
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.PageNum, validation.Type("int64"),
			validation.GreaterEqual(1), validation.Required()),
		validation.Parameter(&params.PageSize, validation.Type("int64"),
			validation.GreaterEqual(1), validation.LessEqual(50), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)
	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候
	parentFolderId, err := folderDbx.GetParentFolderId(params.AppId, p, false, params.UserId)
	if err != nil || parentFolderId == primitive.NilObjectID {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	getFolders, err := folderDbx.GetFolderTreeByParentFolderId(params.AppId, parentFolderId, []int64{1, 0, -1, -2})
	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	parentFolderIdList := []primitive.ObjectID{
		parentFolderId,
	}

	pathMap := map[primitive.ObjectID]string{
		parentFolderId: params.Path,
	}

	for _, v := range getFolders {
		pathMap[v.Id] = path.Join(pathMap[v.ParentFolderId], v.FolderName)
		parentFolderIdList = append(parentFolderIdList, v.Id)
	}

	// 4、操作数据库
	files, err := fileDbx.GetFileLisByParentFolderIdList(
		params.AppId,
		parentFolderIdList,
		params.PageNum,
		params.PageSize,
		[]int64{-1})
	// files, err := fileDbx.GetFileLisByAuthorId(
	// 	params.AppId,
	// 	params.UserId,
	// 	params.PageNum,
	// 	params.PageSize,
	// 	[]int64{-1})

	if err != nil {
		res.Error = err.Error()
		res.Code = 10006
		res.Call(c)
		return
	}

	res.Code = 200
	// log.Info("file", file)
	tempList := []map[string]interface{}{}

	if len(files) == 0 {
		res.Data = map[string]interface{}{
			"total": 0,
			"list":  len(files),
		}
		res.Call(c)
		return
	}

	hashList := []string{}
	for _, v := range files {
		hashList = append(hashList, v.Hash)
	}

	staticFileList, err := fileDbx.GetStaticFileListWithHash(hashList)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	for _, v := range files {
		for _, sv := range staticFileList {
			if sv.FileInfo.Hash == v.Hash {
				// pd := ""
				// if v.AvailableRange.Password != "" {
				// 	pd = v.AvailableRange.Password[0:2] + "******" + v.AvailableRange.Password[len(v.AvailableRange.Password)-3:len(v.AvailableRange.Password)]
				// }
				tempList = append(tempList, map[string]interface{}{
					"id":             v.Id,
					"shortId":        v.ShortId,
					"fileName":       v.FileName,
					"path":           pathMap[v.ParentFolderId],
					"parentFolderId": v.ParentFolderId.Hex(),
					// "fileInfo":       map[string]interface{}{
					// 	"name"
					// },
					"availableRange": map[string]interface{}{
						"visitCount":     v.AvailableRange.VisitCount,
						"expirationTime": v.AvailableRange.ExpirationTime,
						"password":       v.AvailableRange.Password,
						"allowShare":     v.AvailableRange.AllowShare,
						"shareUsers":     v.AvailableRange.ShareUsers,
						"authorId":       v.AvailableRange.AuthorId,
					},
					"usage": map[string]interface{}{
						"visitCount": v.Usage.VisitCount,
					},
					"createTime":     v.CreateTime,
					"lastUpdateTime": v.LastUpdateTime,
					"deleteTime":     v.DeleteTime,
					"urls": map[string]string{
						"domainUrl": conf.Config.StaticPathDomain,
						"shortUrl":  "/s/" + v.ShortId,
						"url":       "/s" + path.Join(pathMap[v.ParentFolderId], v.FileName) + "?a=" + conf.AppList[v.AppId].EncryptionId + "&r=" + params.RootPath,
					},
					"fileInfo": map[string]interface{}{
						"name":         sv.FileInfo.Name,
						"size":         sv.FileInfo.Size,
						"type":         sv.FileInfo.Type,
						"suffix":       sv.FileInfo.Suffix,
						"lastModified": sv.FileInfo.LastModified,
						"hash":         sv.FileInfo.Hash,
						"width":        sv.FileInfo.Width,
						"height":       sv.FileInfo.Height,
					},
				})
				break
			}
		}
	}
	res.Data = map[string]interface{}{
		"total": len(files),
		"list":  tempList,
	}
	res.Call(c)
}

func (dc *FileController) FilterFile(file *models.File) error {
	// 检测访问次数
	if file.AvailableRange.VisitCount != -1 && file.Usage.VisitCount > file.AvailableRange.VisitCount {
		fileDbx.FileNotAccessible(file.Id)
		return errors.New("404")
	}
	if file.AvailableRange.ExpirationTime != -1 && file.AvailableRange.ExpirationTime <= time.Now().Unix() {
		// 已过期
		fileDbx.ExpiredFile(file.Id)
		return errors.New("404")
	}
	// 检测有效期
	return nil
}

func (dc *FileController) VisitFile(file *models.File) error {
	if err := fileDbx.VisitFile(file.Id); err != nil {
		return errors.New("404")
	}
	return nil
}

func (dc *FileController) GetPasswordToken(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId    string
		AppKey   string
		UserId   string
		Path     string
		FileName string
		RootPath string
		Deadline int64
	}{
		AppId:    c.GetString("appId"),
		AppKey:   c.GetString("appKey"),
		UserId:   c.GetString("userId"),
		Path:     c.PostForm("path"),
		FileName: c.PostForm("fileName"),
		RootPath: c.PostForm("rootPath"),
		Deadline: nint.ToInt64(c.PostForm("deadline")),
	}
	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Deadline, validation.Type("int64"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	file, err := fileDbx.GetFileWithFileInfo(params.AppId, path.Join(params.RootPath, params.Path), params.FileName, params.UserId)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	t := time.Duration(params.Deadline-time.Now().Unix()) * time.Second

	u, p := ncredentials.GenerateCredentials(params.AppKey+file.AvailableRange.Password, t)
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	// _, r := ncredentials.GenerateCredentials(params.AppKey+params.RootPath, t)
	// if err != nil {
	// 	res.Errors(err)
	// 	res.Code = 10006
	// 	res.Call(c)
	// 	return
	// }

	res.Code = 200
	res.Data = map[string]interface{}{
		"user":          u,
		"passwordToken": p,
		// "rootPathToken": r,
	}
	res.Call(c)
}

func (dc *FileController) ProcessFile(c *gin.Context, filePath string) (string, error) {
	process := c.Query("x-saass-process")

	processSplit := strings.Split(process, ",")
	processType := processSplit[0]

	fileNameWithSuffix := path.Base(filePath)
	fileType := path.Ext(fileNameWithSuffix)
	fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)

	switch processType {
	case "image/resize":
		if !(fileType == ".jpg" || fileType == ".jpeg" || fileType == ".png") {
			return "", errors.New("this is not a picture")
		}
		imageInfo, err := nimages.GetImageInfo(filePath)
		if err != nil {
			return "", err
		}
		quality := nint.ToInt(processSplit[2])
		pixel := nint.ToInt64(processSplit[1])

		saveAsFoloderPath := "./static/temp/" + strings.Replace(
			strings.Replace(filePath, fileNameOnly+fileType, "", -1), "./static/storage/", "", -1)
		saveAsPath := fileNameOnly + "_" + processSplit[1] + "_" + processSplit[2] + fileType

		// log.Info("saveAsPath", saveAsFoloderPath, saveAsPath)
		// 创建文件夹
		if !nfile.IsExists(saveAsFoloderPath) {
			os.MkdirAll(saveAsFoloderPath, os.ModePerm)
		}
		if !nfile.IsExists(saveAsFoloderPath + saveAsPath) {
			w := 0
			h := 0
			if imageInfo.Width > imageInfo.Height {
				if pixel > imageInfo.Width {
					w = nint.ToInt(imageInfo.Width)
				} else {
					w = nint.ToInt(pixel)
				}
			} else {
				if pixel > imageInfo.Height {
					h = nint.ToInt(imageInfo.Height)
				} else {
					h = nint.ToInt(pixel)
				}
			}
			err = nimages.Resize(filePath, saveAsFoloderPath+saveAsPath, w, h, quality)
			if err != nil {
				return "", err
			}
		}
		return saveAsFoloderPath + saveAsPath, nil
	default:
		return filePath, nil
	}
	// nimages.Resize("./static/WX20210127-125442@2x.png", "./static/3.png", 900, 0, 50)

	// return filePath, nil
}

// 案例
// http://localhost:16100/s/87f7a38b0cdf04949f770ea39264db33?x-saass-process=image/resize,900,70
// http://localhost:16100/s/87f7a38b0cdf04949f770ea39264db33
// http://localhost:16100/s/ces.jpg?a=bc886e5df63bf360077df1f61473e900&x-saass-process=image/resize,200,70
func (dc *FileController) Download(c *gin.Context) {
	// log.Info("------DownloadController------")
	p := c.Request.URL.Path[2:len(c.Request.URL.Path)]
	filePath := p[strings.LastIndex(p, "/")+1 : len(p)-0]
	folderPath := p[0 : strings.LastIndex(p, "/")+1]
	appEncryptionId := c.Query("a")
	sid := c.Query("sid")
	rootPath := c.Query("r")
	userToken := c.Query("ut")
	temporaryAccessToken := c.Query("tat")
	isTAT := false
	shortId := filePath
	if sid != "" {
		shortId = sid
	}
	if temporaryAccessToken != "" {
		u := c.Query("u")
		isTAT = ncredentials.AuthCredentials(u, temporaryAccessToken, shortId)
	}

	log.Info(folderPath, filePath, appEncryptionId == "")
	var file *models.File
	var err error
	appId := ""
	appKey := ""

	log.Info("sid", sid)
	if (folderPath == "/" && appEncryptionId == "") || shortId != "" {

		// log.Info("shortId", shortId)
		file, err = fileDbx.GetFileWithShortId(shortId)
		// log.Info("file", file)

		if file == nil || err != nil {
			c.String(http.StatusNotFound, "")
			return
		}
		for _, v := range conf.AppList {
			if v.AppId == file.AppId {
				appKey = v.AppKey
				break
			}
		}

	} else {
		for _, v := range conf.AppList {
			if v.EncryptionId == appEncryptionId {
				appId = v.AppId
				appKey = v.AppKey
				break
			}
		}
		// log.Info("filePath", path, filePath, 2, strings.LastIndex(path, "/"))
		log.Info("folderPath", folderPath)
		// log.Info(appId, nstrings.StringOr(folderPath, "/"), filePath)
		file, err = fileDbx.GetFileWithFileInfo(appId, path.Join(rootPath, nstrings.StringOr(folderPath, "/")), filePath, "")
		// log.Info("file, err", file, err)
		if file == nil || err != nil {
			c.String(http.StatusNotFound, "")
			return
		}
	}
	log.Info(file.AvailableRange.AllowShare, isTAT)
	if !isTAT {
		switch file.AvailableRange.AllowShare {
		case 1:
			isAll := false
			for _, v := range file.AvailableRange.ShareUsers {
				if v.Uid == "AllUser" {
					isAll = true
					break
				}
			}
			if !isAll {
				c.String(http.StatusNotFound, "")
				return
			}
		case -1:
			// c.String(http.StatusNotFound, "")
			rKey := conf.Redisdb.GetKey("UserToken")
			v, err := conf.Redisdb.Get(rKey.GetKey(userToken))
			if err != nil {
				c.String(http.StatusNotFound, "")
				return
			}
			userId := v.String()
			if userId != file.AvailableRange.AuthorId {
				c.String(http.StatusNotFound, "")
				return
			}
		}
	}
	log.Info("file.AvailableRange.Password ", file.AvailableRange.Password)
	if !isTAT && file.AvailableRange.Password != "" {
		u := c.Query("u")
		p := c.Query("p")
		if u == "" || p == "" {
			c.String(http.StatusNotFound, "")
			return
		}
		// log.Info(appKey + file.AvailableRange.Password)
		// log.Info(u, p, ncredentials.AuthCredentials(u, p, appKey+file.AvailableRange.Password))
		if !ncredentials.AuthCredentials(u, p, appKey+file.AvailableRange.Password) {
			c.String(http.StatusNotFound, "")
			return
		}
	}
	if err = dc.FilterFile(file); err != nil {
		c.String(http.StatusNotFound, "")
		return
	}
	if err = dc.VisitFile(file); err != nil {
		c.String(http.StatusNotFound, "")
		return
	}
	sf, err := fileDbx.GetStaticFileWithHash(file.Hash)
	log.Info(sf)
	if err != nil || sf == nil {
		c.String(http.StatusNotFound, "")
		return
	}
	processFilePath, err := dc.ProcessFile(c, sf.Path+"/"+sf.FileName)
	// log.Info("processFilePath", processFilePath, err)
	if err != nil {
		c.String(http.StatusNotFound, "")
		return
	}
	c.File(processFilePath)
}
