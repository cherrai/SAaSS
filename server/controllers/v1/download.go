package controllersV1

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/nyanyago-utils/ncredentials"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DownloadController struct {
}

func (dc *DownloadController) GetShareFilesHtml(c *gin.Context) {
	log.Info("GetShareFilesHtml")
	var res response.ResponseType
	res.Code = 200
	c.Set("isStaticServer", true)
	// res.Call(c)
	path, _ := os.Executable()
	c.File(filepath.Join(
		strings.Replace(filepath.Dir(path), "tmp", "", -1),
		"./web/share.html"))
}

// http://192.168.204.129:16100/api/v1/share?path=/%E5%AE%89%E8%A3%85%E5%8C%85/meow-backups&sid=EIB8v2Lluy&pwd=2ee422
func (dc *DownloadController) GetShareFiles(c *gin.Context) {
	log.Info("GetShareFiles")
	var res response.ResponseType
	res.Code = 200
	params := struct {
		Sid  string
		Path string
		Pwd  string
	}{
		Sid:  c.Query("sid"),
		Path: filepath.Join(c.Query("path"), "/"),
		Pwd:  c.Query("pwd"),
	}

	var err error
	// 3、验证参数

	if err = validation.ValidateStruct(
		&params,
		validation.Parameter(&params.Sid, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Errors(err)
		res.Code = 10002
		res.Call(c)
		return
	}

	// 先检测文件是否存在
	file, err := fileDbx.GetFileWithShortId(params.Sid)
	if err != nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}
	log.Info(file)
	if file != nil {
		if file.AvailableRange.AllowShare != 1 {
			c.String(http.StatusNotFound, "")
			return
		}
		if file.AvailableRange.Password != "" && file.AvailableRange.Password != params.Pwd {

			c.String(http.StatusNotFound, "")
			return
		}

		list, err := methods.FormatFile([]*models.File{file})
		if err != nil {
			c.String(http.StatusNotFound, "")
			return
		}

		res.Data = map[string]interface{}{
			"list":  list,
			"total": 1,
		}
		res.Call(c)
		return
	}
	// 这里检测文件夹
	folder, err := folderDbx.GetFolderWithShortId(params.Sid)
	if err != nil || folder == nil {
		res.Errors(err)
		res.Code = 10006
		res.Call(c)
		return
	}

	if folder.AvailableRange.AllowShare != 1 {
		c.String(http.StatusNotFound, "")
		return
	}
	if folder.AvailableRange.Password != "" && folder.AvailableRange.Password != params.Pwd {
		c.String(http.StatusNotFound, "")
		return
	}
	log.Info("folder", folder)
	if params.Path == "/" {
		list, err := methods.FormatFolder("/api/v1/share", params.Sid, params.Pwd, params.Path, []*models.Folder{folder})
		if err != nil {
			c.String(http.StatusNotFound, "")
			return
		}

		res.Data = map[string]interface{}{
			"list":  list,
			"total": 1,
		}
		res.Call(c)
		return
	} else {
		pid, err := folderDbx.GetParentFolderIdByPathAndSid(folder.AppId, params.Path, folder.ParentFolderId)
		if err != nil || pid == primitive.NilObjectID {
			c.String(http.StatusNotFound, "")
			return
		}

		folders, err := folderDbx.GetFolderListByParentFolderId(folder.AppId, pid, []int64{1, 0})
		if err != nil {
			res.Errors(err)
			res.Code = 10006
			res.Call(c)
			return
		}
		folderListMap, err := methods.FormatFolder("/api/v1/share", params.Sid, params.Pwd, params.Path, folders)
		if err != nil {
			c.String(http.StatusNotFound, "")
			return
		}

		// 检测是否有文件

		fileList, err := fileDbx.GetFileLisByParentFolderId(folder.AppId, pid)
		if err != nil {
			res.Errors(err)
			res.Code = 10006
			res.Call(c)
			return
		}
		fileListMap, err := methods.FormatFile(fileList.List)
		if err != nil {
			c.String(http.StatusNotFound, "")
			return
		}

		list := append(folderListMap, fileListMap...)
		res.Data = map[string]interface{}{
			"list":  list,
			"total": len(list),
		}

		// res.Data = []map[string]interface{}{

		// 	{
		// 		"id":       "62112a63c88e427c946ff657",
		// 		"category": "electron",
		// 		"name":     "11.5.0/",
		// 		"date":     "2021-08-31T19:21:42Z",
		// 		"type":     "dir",
		// 		"url":      "https://registry.npmmirror.com/-/binary/electron/11.5.0/",
		// 		"modified": "2022-02-19T17:35:31.638Z",
		// 	},
		// }
		// c.JSON(http.StatusOK, res.Data)
		// return
	}

	res.Call(c)
}

func (dc *DownloadController) ProcessFile(c *gin.Context, filePath string) (string, error) {
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
		// imageInfo, err := nimages.GetImageInfoByPath(filePath)
		// if err != nil {
		// 	return "", err
		// }
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

			if err := nimages.ResizeByPath(filePath, saveAsFoloderPath+saveAsPath, nint.ToInt(pixel), 0, 0, quality); err != nil {
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

func (dc *DownloadController) FilterFile(file *models.File) error {
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

func (dc *DownloadController) VisitFile(file *models.File) error {
	if err := fileDbx.VisitFile(file.Id); err != nil {
		return errors.New("404")
	}
	return nil
}

// 案例
// http://localhost:16100/s/87f7a38b0cdf04949f770ea39264db33?x-saass-process=image/resize,900,70
// http://localhost:16100/s/87f7a38b0cdf04949f770ea39264db33
// http://localhost:16100/s/ces.jpg?a=bc886e5df63bf360077df1f61473e900&x-saass-process=image/resize,200,70
func (dc *DownloadController) Download(c *gin.Context) {
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
	shortId := nstrings.StringOr(sid, filePath)
	if temporaryAccessToken != "" {
		u := c.Query("u")
		isTAT = ncredentials.AuthCredentials(u, temporaryAccessToken, shortId)
	}

	log.Info(folderPath, filePath, appEncryptionId == "")
	var file *models.File
	var err error
	appId := ""
	appKey := ""

	if (folderPath == "/" && appEncryptionId == "") || sid != "" {
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
