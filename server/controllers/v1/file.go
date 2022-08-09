package controllersV1

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
)

type FileController struct {
}

func (dc *FileController) DeleteFile(c *gin.Context) {
	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数
	appId := c.PostForm("appId")
	path := c.PostForm("path")
	fileName := c.PostForm("fileName")
	deadlineInRecycleBin := c.PostForm("deadlineInRecycleBin")

	params := struct {
		AppId                string
		Path                 string
		FileName             string
		DeadlineInRecycleBin int64
	}{
		AppId:                appId,
		Path:                 path,
		FileName:             fileName,
		DeadlineInRecycleBin: nint.ToInt64(deadlineInRecycleBin),
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileName, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.DeadlineInRecycleBin, validation.Type("int64"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库

	if err := fileDbx.DeleteFile(params.AppId, params.Path, params.FileName, params.DeadlineInRecycleBin); err != nil {
		res.Error = err.Error()
		res.Code = 10002
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
	appId := c.Query("appId")
	path := c.Query("path")
	fileName := c.Query("fileName")

	params := struct {
		AppId    string
		Path     string
		FileName string
	}{
		AppId:    appId,
		Path:     path,
		FileName: fileName,
	}

	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.FileName, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	file, err := fileDbx.GetFileWithFileInfo(params.AppId, params.Path, params.FileName)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	res.Code = 200
	res.Data = map[string]interface{}{
		"urls": map[string]string{
			"domainUrl":     conf.Config.StaticPathDomain,
			"encryptionUrl": "/s/" + file.EncryptionName,
			"url":           "/s" + file.Path + file.FileName + "?a=" + conf.AppList[appId].EncryptionId,
		},
	}
	res.Call(c)
}
func (dc *FileController) GetFolderFiles(c *gin.Context) {

	// 1、 创建请求体
	var res response.ResponseType
	res.Code = 200

	// 2、获取参数
	appId := c.Query("appId")
	path := c.Query("path")

	params := struct {
		AppId string
		Path  string
	}{
		AppId: appId,
		Path:  path,
	}

	log.Info("params", params)
	// 3、校验参数
	if err := validation.ValidateStruct(
		&params,
		validation.Parameter(&params.AppId, validation.Type("string"), validation.Required()),
		validation.Parameter(&params.Path, validation.Type("string"), validation.Required()),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	// 4、操作数据库
	file, err := fileDbx.GetFileListByPath(params.AppId, params.Path)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}

	res.Code = 200
	log.Info("file", file)
	tempList := []map[string]interface{}{}

	for _, v := range file.List {
		tempList = append(tempList, map[string]interface{}{
			"encryptionName": v.EncryptionName,
			"fileName":       v.FileName,
			"path":           v.Path,
			// "fileInfo":       map[string]interface{}{
			// 	"name"
			// },
			"availableRange": map[string]interface{}{
				"visitCount":     v.AvailableRange.VisitCount,
				"expirationTime": v.AvailableRange.ExpirationTime,
			},
			"usage": map[string]interface{}{
				"visitCount": v.Usage.VisitCount,
			},
			"createTime": v.CreateTime,
			"updateTime": v.UpdateTime,
			"urls": map[string]string{
				"domainUrl":     conf.Config.StaticPathDomain,
				"encryptionUrl": "/s/" + v.EncryptionName,
				"url":           "/s" + v.Path + v.FileName + "?a=" + conf.AppList[v.AppId].EncryptionId,
			},
		})
	}

	res.Data = map[string]interface{}{
		"total": file.Total[0].Count,
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
	path := c.Request.URL.Path[2:len(c.Request.URL.Path)]
	filePath := path[strings.LastIndex(path, "/")+1 : len(path)-0]
	folderPath := path[0 : strings.LastIndex(path, "/")+1]
	encryptionId := c.Query("a")

	if folderPath == "/" && encryptionId == "" {
		encryptionName := filePath
		// log.Info("encryptionName", encryptionName)
		file, err := fileDbx.GetFileWithEncryptionName(encryptionName)
		// log.Info("file", file)
		if file == nil || err != nil {
			c.String(http.StatusNotFound, "")
			return
		}

		if err = dc.FilterFile(file); err != nil {
			c.String(http.StatusNotFound, "")
			return
		}
		if err = dc.VisitFile(file); err != nil {
			c.String(http.StatusNotFound, "")
			return
		}
		// log.Info("hash", file.Hash)
		sf, err := fileDbx.GetStaticFileWithHash(file.Hash)
		// log.Info("sf", sf, err)
		if err != nil || sf == nil {
			c.String(http.StatusNotFound, "")
			return
		}
		processFilePath, err := dc.ProcessFile(c, sf.Path+"/"+sf.FileName)
		// log.Info("processFilePath", processFilePath)
		if err != nil {
			c.String(http.StatusNotFound, "")
			return
		}
		c.File(processFilePath)
		return
	}

	appId := ""
	for _, v := range conf.AppList {
		if v.EncryptionId == encryptionId {
			appId = v.AppId
			break
		}
	}
	// log.Info("filePath", path, filePath, 2, strings.LastIndex(path, "/"))
	// log.Info("folderPath", folderPath)
	// log.Info(appId, nstrings.StringOr(folderPath, "/"), filePath)
	file, err := fileDbx.GetFileWithFileInfo(appId, nstrings.StringOr(folderPath, "/"), filePath)
	// log.Info("file, err", file, err)
	if file == nil || err != nil {
		c.String(http.StatusNotFound, "")
		return
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
	if err != nil || sf == nil {
		c.String(http.StatusNotFound, "")
		return
	}
	processFilePath, err := dc.ProcessFile(c, sf.Path+sf.FileName)
	if err != nil {
		c.String(http.StatusNotFound, "")
		return
	}
	c.File(processFilePath)
}
