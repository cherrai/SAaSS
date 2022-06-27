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
	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/gin-gonic/gin"
)

type DownloadController struct {
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
			return "", errors.New("this is not a picture.")
		}
		imageInfo, err := nimages.GetImageInfo(filePath)
		if err != nil {
			return "", err
		}
		quality := nint.ToInt(processSplit[2])
		pixel := nint.ToInt64(processSplit[1])

		saveAsFoloderPath := "./static/storage/temp/" + strings.Replace(
			strings.Replace(filePath, fileNameOnly+fileType, "", -1), "./static/storage/", "", -1)
		saveAsPath := fileNameOnly + "_" + processSplit[1] + "_" + processSplit[2] + fileType

		// 创建文件夹
		if !methods.IsExists(saveAsFoloderPath) {
			os.MkdirAll(saveAsFoloderPath, os.ModePerm)
		}
		if !methods.IsExists(saveAsFoloderPath + saveAsPath) {
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
func (dc *DownloadController) Download(c *gin.Context) {
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
		processFilePath, err := dc.ProcessFile(c, file.StaticFolderPath+file.StaticFileName)
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
	log.Info(appId, nstrings.StringOr(folderPath, "/"), filePath)
	file, err := fileDbx.GetFileWithFileInfo(appId, nstrings.StringOr(folderPath, "/"), filePath)
	log.Info("file, err", file, err)
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
	processFilePath, err := dc.ProcessFile(c, file.StaticFolderPath+file.StaticFileName)
	if err != nil {
		c.String(http.StatusNotFound, "")
		return
	}
	c.File(processFilePath)
}
