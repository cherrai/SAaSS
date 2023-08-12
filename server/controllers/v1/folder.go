package controllersV1

import (
	"encoding/json"
	"net/http"
	"path"
	"time"

	dbxV1 "github.com/cherrai/SAaSS/dbx/v1"
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/ncredentials"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	folderDbx = dbxV1.FolderDbx{}
)

type FolderController struct {
}

func (dc *FolderController) NewFolder(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId      string
		RootPath   string
		UserId     string
		FolderName string
		ParentPath string
	}{
		AppId:      c.GetString("appId"),
		FolderName: c.PostForm("folderName"),
		UserId:     c.GetString("userId"),
		ParentPath: c.PostForm("parentPath"),
		RootPath:   c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.FolderName, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.ParentPath = path.Join("/", data.ParentPath)
	data.RootPath = path.Join("/", data.RootPath)

	log.Info(data)
	p := path.Join(data.RootPath, data.ParentPath)
	newFolder, err := folderDbx.NewFolder(data.AppId, data.FolderName, p, data.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	res.Data = map[string]string{
		"folderName": newFolder.FolderName,
		// "parentPath": newFolder.ParentPath,
	}
	res.Call(c)
}

func (dc *FolderController) GetRootFolderToken(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	params := struct {
		AppId    string
		AppKey   string
		RootPath string
		UserId   string
		Deadline int64
	}{
		AppId:    c.GetString("appId"),
		AppKey:   c.GetString("appKey"),
		UserId:   c.GetString("userId"),
		RootPath: c.PostForm("rootPath"),
		Deadline: nint.ToInt64(c.PostForm("deadline")),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		params.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Deadline, validation.Type("int64"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	t := time.Duration(params.Deadline-time.Now().Unix()) * time.Second

	u, p := ncredentials.GenerateCredentials(params.AppKey+params.RootPath, t)
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}

	res.Code = 200
	res.Data = map[string]interface{}{
		"user":          u,
		"rootPathToken": p,
	}
	res.Call(c)
}

func (dc *FolderController) GetFolderByShortId(c *gin.Context) {
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
		validation.Parameter(&params.Id, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	// 4、操作数据库
	v, err := folderDbx.GetFolderWithShortId(params.Id)
	if err != nil {
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
			"id":         v.Id,
			"shortId":    v.ShortId,
			"folderName": v.FolderName,
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
		}

		res.Data = map[string]interface{}{
			"id":             v.Id,
			"folderName":     v.FolderName,
			"shortId":        v.ShortId,
			"parentFolderId": v.ParentFolderId,
			"status":         v.Status,
			"availableRange": map[string]interface{}{
				"password":   v.AvailableRange.Password,
				"allowShare": v.AvailableRange.AllowShare,
				"shareUsers": v.AvailableRange.ShareUsers,
				"authorId":   v.AvailableRange.AuthorId,
			},
			"usage":          map[string]interface{}{},
			"createTime":     v.CreateTime,
			"lastUpdateTime": v.LastUpdateTime,
			"deleteTime":     v.DeleteTime,
			"accessToken":    methods.GetTemporaryAccessToken(v.ShortId, params.Deadline),
		}
	}

	res.Code = 200
	res.Call(c)
}

func (dc *FolderController) GetFolderListWithShortId(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200
	params := struct {
		AppId       string
		AppKey      string
		UserId      string
		Id          string
		AccessToken map[string]string
		Deadline    int64
		Path        int64
	}{
		AppId:       c.GetString("appId"),
		AppKey:      c.GetString("appKey"),
		UserId:      c.GetString("userId"),
		Id:          c.Query("id"),
		AccessToken: c.QueryMap("accessToken"),
		Deadline:    nint.ToInt64(c.Query("deadline")),
		// Path:        c.Query("path"),
	}
	at := c.Query("accessToken")
	if at != "" {
		err := json.Unmarshal([]byte(at), &params.AccessToken)
		if err != nil {
			res.Errors(err)
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
	folders, err := folderDbx.GetFolderListByParentFolderId(folder.AppId, folder.Id, []int64{1})
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	list := []map[string]interface{}{}
	for _, v := range folders {
		list = append(list, map[string]interface{}{
			"id":             v.Id,
			"folderName":     v.FolderName,
			"shortId":        v.ShortId,
			"parentFolderId": v.ParentFolderId,
			"status":         v.Status,
			"availableRange": map[string]interface{}{
				"password":   v.AvailableRange.Password,
				"allowShare": v.AvailableRange.AllowShare,
				"shareUsers": v.AvailableRange.ShareUsers,
				"authorId":   v.AvailableRange.AuthorId,
			},
			"usage":          map[string]interface{}{},
			"createTime":     v.CreateTime,
			"lastUpdateTime": v.LastUpdateTime,
			"deleteTime":     v.DeleteTime,
			"accessToken":    methods.GetTemporaryAccessToken(v.ShortId, params.Deadline),
		})
	}
	res.Data = map[string]interface{}{
		"list":  list,
		"total": len(folders),
	}
	res.Call(c)
}

func (dc *FolderController) GerFolderList(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId      string
		AppKey     string
		UserId     string
		RootPath   string
		ParentPath string
	}{
		AppId:      c.GetString("appId"),
		AppKey:     c.GetString("appKey"),
		UserId:     c.GetString("userId"),
		ParentPath: c.Query("parentPath"),
		RootPath:   c.Query("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.AppKey, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.ParentPath = path.Join("/", data.ParentPath)
	data.RootPath = path.Join("/", data.RootPath)

	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候

	parentFolderId, err := folderDbx.GetParentFolderId(data.AppId,
		path.Join(data.RootPath, data.ParentPath),
		false,
		data.UserId)

	log.Info(parentFolderId)
	if err != nil || parentFolderId == primitive.NilObjectID {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	folders, err := folderDbx.GetFolderListByParentFolderId(data.AppId, parentFolderId, []int64{1, 0})
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	list := []map[string]interface{}{}
	for _, v := range folders {
		list = append(list, map[string]interface{}{
			"id":             v.Id,
			"folderName":     v.FolderName,
			"shortId":        v.ShortId,
			"parentFolderId": v.ParentFolderId,
			"path":           data.ParentPath,
			"status":         v.Status,
			"availableRange": map[string]interface{}{
				"password":   v.AvailableRange.Password,
				"allowShare": v.AvailableRange.AllowShare,
				"shareUsers": v.AvailableRange.ShareUsers,
				"authorId":   v.AvailableRange.AuthorId,
			},
			"usage":          map[string]interface{}{},
			"createTime":     v.CreateTime,
			"lastUpdateTime": v.LastUpdateTime,
			"deleteTime":     v.DeleteTime,
		})
	}
	res.Data = map[string]interface{}{
		"list":  list,
		"total": len(folders),
	}
	res.Call(c)
}

func (dc *FolderController) RenameFolder(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId         string
		AppKey        string
		RootPath      string
		UserId        string
		NewFolderName string
		OldFolderName string
		ParentPath    string
	}{
		AppId:         c.GetString("appId"),
		AppKey:        c.GetString("appKey"),
		UserId:        c.GetString("userId"),
		NewFolderName: c.PostForm("newFolderName"),
		OldFolderName: c.PostForm("oldFolderName"),
		ParentPath:    c.PostForm("parentPath"),
		RootPath:      c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.AppKey, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.NewFolderName, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.OldFolderName, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.RootPath = path.Join("/", data.RootPath)
	p := path.Join(data.RootPath, data.ParentPath)
	getFolder, err := folderDbx.GetFolder(data.AppId, data.NewFolderName, p, data.UserId)
	log.Info(getFolder, err, err != nil || getFolder != nil)
	if err != nil || getFolder != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	err = folderDbx.UpdateFolderName(data.AppId, p, data.OldFolderName, data.NewFolderName, data.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	res.Data = map[string]string{}
	res.Call(c)
}

func (dc *FolderController) MoveFoldersToTrash(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId       string
		RootPath    string
		UserId      string
		ParentPath  string
		FolderNames map[string]string
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		ParentPath:  c.PostForm("parentPath"),
		FolderNames: c.PostFormMap("folderNames"),
		RootPath:    c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.FolderNames, validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.RootPath = path.Join("/", data.RootPath)

	folderNames := []string{}
	for _, v := range data.FolderNames {
		folderNames = append(folderNames, v)
	}
	log.Info("folderNames", folderNames)
	err = folderDbx.MoveFoldersToTrash(data.AppId, path.Join(data.RootPath, data.ParentPath), folderNames, data.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	res.Data = map[string]string{}
	res.Call(c)
}

func (dc *FolderController) CheckFolderExists(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId       string
		RootPath    string
		UserId      string
		ParentPath  string
		FolderNames map[string]string
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		ParentPath:  c.PostForm("parentPath"),
		FolderNames: c.PostFormMap("folderNames"),
		RootPath:    c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.FolderNames, validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.RootPath = path.Join("/", data.RootPath)

	folderNames := []string{}
	for _, v := range data.FolderNames {
		folderNames = append(folderNames, v)
	}
	log.Info("folderNames", folderNames)
	folders, err := folderDbx.GetFolderListByFolderNames(data.AppId, path.Join(data.RootPath, data.ParentPath), folderNames, data.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10001
		res.Call(c)
		return
	}
	list := []string{}
	for _, v := range folders {
		list = append(list, v.FolderName)
	}
	res.Data = map[string]interface{}{
		"list":  list,
		"total": len(list),
	}
	res.Call(c)
}

func (dc *FolderController) RestoreFolder(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId       string
		RootPath    string
		UserId      string
		ParentPath  string
		FolderNames map[string]string
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		ParentPath:  c.PostForm("parentPath"),
		FolderNames: c.PostFormMap("folderNames"),
		RootPath:    c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.FolderNames, validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.RootPath = path.Join("/", data.RootPath)

	folderNames := []string{}
	for _, v := range data.FolderNames {
		folderNames = append(folderNames, v)
	}
	log.Info("folderNames", folderNames)
	err = folderDbx.Restore(data.AppId, path.Join(data.RootPath, data.ParentPath), folderNames, data.UserId)
	if err != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	res.Data = map[string]string{}
	res.Call(c)
}

func (dc *FolderController) DeleteFolders(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId       string
		RootPath    string
		UserId      string
		ParentPath  string
		FolderNames map[string]string
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		ParentPath:  c.PostForm("parentPath"),
		FolderNames: c.PostFormMap("folderNames"),
		RootPath:    c.PostForm("rootPath"),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.FolderNames, validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.RootPath = path.Join("/", data.RootPath)

	folderNames := []string{}
	for _, v := range data.FolderNames {
		folderNames = append(folderNames, v)
	}
	log.Info("folderNames", folderNames)
	err = folderDbx.DeleteFolders(data.AppId,
		data.UserId,
		path.Join(data.RootPath, data.ParentPath),
		folderNames)
	if err != nil {
		res.Errors(err)
		res.Code = 10020
		res.Call(c)
		return
	}
	res.Data = map[string]string{}
	res.Call(c)
}

func (dc *FolderController) SetFolderSharing(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId       string
		UserId      string
		RootPath    string
		Path        string
		FolderNames map[string]string
		Status      int64
	}{
		AppId:       c.GetString("appId"),
		UserId:      c.GetString("userId"),
		RootPath:    c.PostForm("rootPath"),
		Path:        c.PostForm("path"),
		FolderNames: c.PostFormMap("folderNames"),
		Status:      nint.ToInt64(c.PostForm("status")),
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
		validation.Parameter(&params.FolderNames, validation.Required()),
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Status, validation.Type("int64"), validation.Enum([]int64{1, -1}), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	folderNames := []string{}
	for _, v := range params.FolderNames {
		folderNames = append(folderNames, v)
	}
	p := path.Join(params.RootPath, params.Path)

	// 4、操作数据库
	if err := folderDbx.SetFileSharing(params.AppId,
		params.UserId,
		p,
		folderNames,
		params.Status); err != nil {
		res.Errors(err)
		res.Code = 10011
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

// 开发中
func (dc *FolderController) SetFolderPassword(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId      string
		UserId     string
		RootPath   string
		Path       string
		FolderName string
		Password   string
	}{
		AppId:      c.GetString("appId"),
		UserId:     c.GetString("userId"),
		RootPath:   c.PostForm("rootPath"),
		Path:       c.PostForm("path"),
		FolderName: c.PostForm("folderName"),
		Password:   c.PostForm("password"),
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
		validation.Parameter(&params.FolderName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Password, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.Path)

	if params.Password == "noPassword" {
		params.Password = ""
	}

	// 4、操作数据库
	if err := folderDbx.SetFolderPassword(params.AppId,
		params.UserId,
		p,
		params.FolderName,
		params.Password); err != nil {
		res.Errors(err)
		res.Code = 10011
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FolderController) CopyFolder(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId         string
		UserId        string
		RootPath      string
		ParentPath    string
		FolderNames   map[string]string
		NewParentPath string
	}{
		AppId:         c.GetString("appId"),
		UserId:        c.GetString("userId"),
		RootPath:      c.PostForm("rootPath"),
		ParentPath:    c.PostForm("parentPath"),
		FolderNames:   c.PostFormMap("folderNames"),
		NewParentPath: c.PostForm("newParentPath"),
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
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FolderNames, validation.Required()),
		validation.Parameter(&params.NewParentPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.ParentPath)
	np := path.Join(params.RootPath, params.NewParentPath)

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FolderNames {
		fns = append(fns, v)
	}
	if err := folderDbx.CopyFolder(params.AppId, params.UserId, p, fns, np); err != nil {
		res.Errors(err)
		res.Code = 10021
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FolderController) MoveFolder(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数

	params := struct {
		AppId         string
		UserId        string
		RootPath      string
		ParentPath    string
		FolderNames   map[string]string
		NewParentPath string
	}{
		AppId:         c.GetString("appId"),
		UserId:        c.GetString("userId"),
		RootPath:      c.PostForm("rootPath"),
		ParentPath:    c.PostForm("parentPath"),
		FolderNames:   c.PostFormMap("folderNames"),
		NewParentPath: c.PostForm("newParentPath"),
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
		validation.Parameter(&params.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.ParentPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FolderNames, validation.Required()),
		validation.Parameter(&params.NewParentPath, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	p := path.Join(params.RootPath, params.ParentPath)
	np := path.Join(params.RootPath, params.NewParentPath)

	// 4、操作数据库
	fns := []string{}
	for _, v := range params.FolderNames {
		fns = append(fns, v)
	}
	if err := folderDbx.MoveFolder(params.AppId, params.UserId, p, fns, np); err != nil {
		log.Info(err)
		res.Errors(err)
		res.Code = 10021
		res.Call(c)
		return
	}
	res.Code = 200
	res.Call(c)
}

func (dc *FolderController) GetRecyclebinFolderList(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200

	data := struct {
		AppId    string
		AppKey   string
		UserId   string
		RootPath string
		Path     string
		PageNum  int64
		PageSize int64
	}{
		AppId:    c.GetString("appId"),
		AppKey:   c.GetString("appKey"),
		UserId:   c.GetString("userId"),
		Path:     c.Query("path"),
		RootPath: c.Query("rootPath"),
		PageNum:  nint.ToInt64(c.Query("pageNum")),
		PageSize: nint.ToInt64(c.Query("pageSize")),
	}

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		data.RootPath = t.RootPath
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&data,
		validation.Parameter(&data.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.AppKey, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.RootPath, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&data.PageNum, validation.Type("int64"),
			validation.GreaterEqual(1), validation.Required()),
		validation.Parameter(&data.PageSize, validation.Type("int64"),
			validation.GreaterEqual(1), validation.LessEqual(50), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}
	data.Path = path.Join("/", data.Path)
	data.RootPath = path.Join("/", data.RootPath)

	// 如果是直接根目录获取，可以检测下行不行，
	// 譬如获取所有目录内容的时候

	parentFolderId, err := folderDbx.GetParentFolderId(data.AppId,
		path.Join(data.RootPath, data.Path),
		false,
		data.UserId)

	if err != nil || parentFolderId == primitive.NilObjectID {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	folders, err := folderDbx.GetFolderTreeByParentFolderId(data.AppId, parentFolderId, []int64{1, 0, -1, -2})
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	// folders, err := folderDbx.GetFolderListByAuthorId(data.AppId, data.UserId, data.PageNum, data.PageSize, []int64{-1})
	// if err != nil {
	// 	res.Errors(err)
	// 	res.Code = 10006
	// 	res.Call(c)
	// 	return
	// }
	pathMap := map[primitive.ObjectID]string{
		parentFolderId: data.Path,
	}
	list := []map[string]interface{}{}
	for _, v := range folders {
		pathMap[v.Id] = path.Join(pathMap[v.ParentFolderId], v.FolderName)
		if v.Status == -1 {
			list = append(list, map[string]interface{}{
				"id":             v.Id,
				"folderName":     v.FolderName,
				"shortId":        v.ShortId,
				"parentFolderId": v.ParentFolderId,
				"path":           pathMap[v.ParentFolderId],
				"status":         v.Status,
				"availableRange": map[string]interface{}{
					"password":   v.AvailableRange.Password,
					"allowShare": v.AvailableRange.AllowShare,
					"shareUsers": v.AvailableRange.ShareUsers,
					"authorId":   v.AvailableRange.AuthorId,
				},
				"usage":          map[string]interface{}{},
				"createTime":     v.CreateTime,
				"lastUpdateTime": v.LastUpdateTime,
				"deleteTime":     v.DeleteTime,
			})
		}
	}
	res.Data = map[string]interface{}{
		"list":  list,
		"total": len(list),
	}
	res.Call(c)
}
