package controllersV1

import (
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/gin-gonic/gin"
)

var tempFolderPath = "./static/csf/"

type CloudServiceFileController struct {
}

func (fc *CloudServiceFileController) GetBasePath(appId string, path string) string {

	return filepath.Join(tempFolderPath, appId, path)
}

func (fc *CloudServiceFileController) GetFilePath(appId string, path string, fileName string) string {
	return filepath.Join(
		fc.GetBasePath(appId, path),
		fileName)
}

func (fc *CloudServiceFileController) GetDeadlineFilePath(appId string, path string, fileName string) string {
	return filepath.Join(
		fc.GetBasePath(appId, path),
		strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))+"_deadline")

}

func (fc *CloudServiceFileController) UploadFile(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200
	log.Info("------UploadFile------")
	// fileConfigInfo := c.MustGet("fileConfigInfo").(*typings.TempFileConfigInfo)
	path := c.PostForm("path")
	deadline := nint.ToInt64(c.PostForm("deadline"))
	appId := c.GetString("appId")

	file, err := c.FormFile("files")
	if err != nil {
		res.Code = 10016
		res.Errors(err)
		res.Call(c)
		return
	}

	// fileInfoMap := make(map[string]string)
	fileName, err := url.QueryUnescape(file.Filename)
	if err != nil {
		res.Code = 10016
		res.Errors(err)
		res.Call(c)
		return
	}

	log.Info(path, fileName, c.PostForm("fileName"))
	if fileName != c.PostForm("fileName") {
		res.Code = 10002
		res.Errors(err)
		res.Call(c)
		return
	}

	log.Info(path, fileName)

	log.Info("GetBasePath", fc.GetBasePath(appId, path))

	if err := nfile.CreateFolder(fc.GetBasePath(appId, path), 0750); err != nil {
		res.Code = 10016
		res.Errors(err)
		res.Call(c)
		return
	}

	if deadline > 0 {
		deadlineFilePath := fc.GetDeadlineFilePath(appId, path, fileName)
		if err = nfile.CreateFile(deadlineFilePath, deadline); err != nil {
			res.Code = 10016
			res.Errors(err)
			res.Call(c)
			return
		}
	}
	filePath := fc.GetFilePath(appId, path, fileName)

	if err = c.SaveUploadedFile(file, filePath); err != nil {
		res.Code = 10016
		res.Errors(err)
		res.Call(c)
		return
	}

	res.Code = 200
	res.Call(c)
}

func (fc *CloudServiceFileController) GetFile(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200
	path := c.Query("path")
	fileName := c.Query("fileName")
	appId := c.GetString("appId")

	filePath := fc.GetFilePath(appId, path, fileName)
	deadlineFilePath := fc.GetDeadlineFilePath(appId, path, fileName)

	deadline := int64(0)

	if nfile.IsExists(deadlineFilePath) {
		if err := nfile.ReadFile(deadlineFilePath, &deadline); err != nil {
			res.Code = 10024
			res.Errors(err)
			res.Call(c)
			return
		}
		if time.Now().Unix() > deadline {
			res.Code = 10024
			res.Call(c)
			if err := nfile.Remove(filePath); err != nil {
				res.Code = 10025
				res.Errors(err)
				res.Call(c)
				return
			}
			if err := nfile.Remove(deadlineFilePath); err != nil {
				res.Code = 10025
				res.Errors(err)
				res.Call(c)
				return
			}
			if err := nfile.RemoveEmptyFolder(fc.GetBasePath(appId, path)); err != nil {
				res.Code = 10025
				res.Errors(err)
				res.Call(c)
				return
			}
			return
		}
	}
	if !nfile.IsExists(filePath) {
		res.Code = 10024
		res.Call(c)
		return
	}

	c.File(filePath)
}
