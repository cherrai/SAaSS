package controllersV1

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	dbxV1 "github.com/cherrai/SAaSS/dbx/v1"
	"github.com/cherrai/SAaSS/models"

	"github.com/cherrai/SAaSS/services/methods"
	"github.com/cherrai/SAaSS/services/response"
	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/cipher"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
)

var (
	log = nlog.New()
)

var fileDbx = new(dbxV1.FileDbx)

type ChunkUploadController struct {
}

// 如果接着上次的传，则需要获取上次的上传进度
func (fc *ChunkUploadController) CreateChunk(c *gin.Context) {
	// log.Info("------CreateChunk------")
	var res response.ResponseType
	res.Code = 200

	appId := c.PostForm("appId")
	// appKey := c.PostForm("appKey")
	// typestr := c.PostForm("type")
	fileName := c.PostForm("fileName")
	fileInfoStr := c.PostForm("fileInfo")
	fileInfo := c.PostFormMap("fileInfo")
	if fileInfoStr != "" {
		err := json.Unmarshal([]byte(fileInfoStr), &fileInfo)
		if err != nil {
			res.Error = err.Error()
			res.Code = 10002
			res.Call(c)
			return
		}
		// fileInfo
	}
	// log.Info("fileName", fileName)
	// log.Info("fileInfo", fileInfo)
	// log.Info("typestr", typestr)
	// 存在则直接返地址，并不设置任何token
	tempFolderPath := "./static/chuck/" + fileInfo["hash"] + "/"

	fileNameWithSuffix := path.Base(fileInfo["name"])
	fileType := path.Ext(fileNameWithSuffix)
	fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)
	fileConfigInfo := typings.TempFileConfigInfo{
		AppId:          appId,
		Name:           fileName,
		EncryptionName: strings.ToLower(cipher.MD5(fileInfo["hash"] + appId + fileInfo["size"] + nstrings.ToString(time.Now().Unix()))),
		Path:           c.PostForm("path"),
		// StaticFolderPath: strings.ToLower("./static/storage" +
		// 	"/" + time.Now().Format("2006/01/02")),
		// StaticFileName:      strings.ToLower(cipher.MD5(fileInfo["hash"]+fileInfo["size"]) + fileType),
		TempFolderPath:      tempFolderPath,
		TempChuckFolderPath: tempFolderPath + "/chuck/",
		// Type:                typestr,
		ChunkSize:      nint.ToInt64(c.PostForm("chunkSize")),
		CreateTime:     time.Now().Unix(),
		ExpirationTime: nint.ToInt64(c.PostForm("expirationTime")),
		VisitCount:     nint.ToInt64(c.PostForm("visitCount")),

		FileInfo: typings.FileInfo{
			Name:         fileNameOnly,
			Size:         nint.ToInt64(fileInfo["size"]),
			Type:         fileInfo["type"],
			Suffix:       fileType,
			LastModified: nint.ToInt64(fileInfo["lastModified"]),
			Hash:         fileInfo["hash"],
		},
	}

	err := validation.ValidateStruct(
		&fileConfigInfo,
		validation.Parameter(&fileConfigInfo.Name, validation.Required()),
		validation.Parameter(&fileConfigInfo.Path, validation.Required()),
		validation.Parameter(&fileConfigInfo.TempFolderPath, validation.Required()),
		validation.Parameter(&fileConfigInfo.TempChuckFolderPath, validation.Required()),
		validation.Parameter(&fileConfigInfo.ChunkSize, validation.Type("int64"), validation.Required(), validation.Enum([]int64{16 * 1024, 32 * 1024, 64 * 1024, 128 * 1024, 256 * 1024, 512 * 1024, 1024 * 1024}), validation.GreaterEqual(1)),
		// validation.Parameter(&fileConfigInfo.Type, validation.Type("string"), validation.Required(), validation.Enum([]string{"Image", "Video", "Audio", "Text", "File"})),
	)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	err = validation.ValidateStruct(
		&fileConfigInfo.FileInfo,
		validation.Parameter(&fileConfigInfo.FileInfo.Name, validation.Required()),
		validation.Parameter(&fileConfigInfo.FileInfo.Size, validation.Type("int64"), validation.Required(), validation.GreaterEqual(1)),
		validation.Parameter(&fileConfigInfo.FileInfo.Type, validation.Required()),
		validation.Parameter(&fileConfigInfo.FileInfo.LastModified, validation.Type("int64"), validation.Required(), validation.GreaterEqual(1)),
		validation.Parameter(&fileConfigInfo.FileInfo.Hash, validation.Required()),
	)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	// 通过检测StaticFileName看数据库文件是否存在

	/**
	1、判断该path和filename的File是否存在
	1.1 存在则判断hash是否一致
	不一致则重新上传，创建心的staticfile.并覆盖之前的hash
	一致则查看hash所属的StaticFile是否还存在
		1.1.1 存在则视为不用再上传了
		不存在则需要上传


	1.1 若File不存在则判断hash的StaticFile是否存在。存在则创建File。
	不存在则重新上传

	*/
	file, err := fileDbx.GetFileWithFileInfo(fileConfigInfo.AppId, fileConfigInfo.Path, fileConfigInfo.Name)
	log.Info("file", file)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10019
		res.Call(c)
		return
	}

	// 内容存在
	staticFilesIsExist := false

	sf, err := fileDbx.GetStaticFileWithHash(fileConfigInfo.FileInfo.Hash)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10019
		res.Call(c)
		return
	}
	if sf != nil {
		// 静态文件存在则直接更新即可
		if nfile.IsExists(sf.Path + "/" + sf.FileName) {
			staticFilesIsExist = true
		}
	}
	if staticFilesIsExist {
		// 内容存在则更新
		if file != nil && file.Hash != "" {
			// 更新内容到最新状态
			if file.Hash != fileConfigInfo.FileInfo.Hash {
				file.HashHistory = append(file.HashHistory, models.HashHistory{
					Hash: file.Hash,
				})
				file.Hash = fileConfigInfo.FileInfo.Hash
			}
			file.Status = 1
			file.DeleteTime = -1
			file.DeadlineInRecycleBin = -1
			file.AvailableRange.VisitCount = fileConfigInfo.VisitCount
			file.AvailableRange.ExpirationTime = fileConfigInfo.ExpirationTime
			fileConfigInfo.EncryptionName = file.EncryptionName

			_, err := fileDbx.UpdateFile(file)
			if err != nil {
				res.Error = err.Error()
				res.Code = 10019
				res.Call(c)
				return
			}
			res.Data = map[string]interface{}{
				"urls": methods.GetResponseData(&fileConfigInfo),
			}
			res.Code = 200
			res.Call(c)
			return
		} else {
			// 内容不存在则创建
			// appId 不一样
			file := models.File{
				AppId:          fileConfigInfo.AppId,
				EncryptionName: fileConfigInfo.EncryptionName,
				FileName:       fileConfigInfo.Name,
				Path:           fileConfigInfo.Path,
				AvailableRange: models.FileAvailableRange{
					VisitCount:     fileConfigInfo.VisitCount,
					ExpirationTime: fileConfigInfo.ExpirationTime,
				},
				Hash: fileConfigInfo.FileInfo.Hash,
			}
			_, err := fileDbx.SaveFile(&file)
			if err != nil {
				res.Code = 10016
				res.Error = err.Error()
				res.Call(c)
				return
			}
			res.Data = map[string]interface{}{
				"urls": methods.GetResponseData(&fileConfigInfo),
			}
			res.Code = 200
			res.Call(c)
			return
		}
	}

	// 创建文件信息的临时配置文件
	if !nfile.IsExists(fileConfigInfo.TempChuckFolderPath) {
		os.MkdirAll(fileConfigInfo.TempChuckFolderPath, os.ModePerm)
	}
	// 保存临时配置文件
	serverConfig, serverConfigErr := os.Create(fileConfigInfo.TempFolderPath + "/info.json")
	defer serverConfig.Close()
	if serverConfigErr != nil {
		res.Error = serverConfigErr.Error()
		res.Code = 10001
		res.Call(c)
		return
	}

	str, _ := json.MarshalIndent(fileConfigInfo, "", "  ")
	_, serverConfigWriteErr := serverConfig.Write([]byte(str))
	if serverConfigWriteErr != nil {
		res.Error = serverConfigErr.Error()
		res.Code = 10001
		res.Call(c)
		return
	}

	// 获取token
	// log.Info("configInfo", configInfo)
	token, err := methods.GetToken(fileConfigInfo)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	// 后续启用redis
	err = conf.Redisdb.Set("file_"+fileInfo["hash"], token, 5*60*time.Second)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	// 获取上次的传输进度
	totalSize, uploadedOffset := methods.GetUploadedOffset(&fileConfigInfo)
	err = conf.Redisdb.Set("file_"+fileInfo["hash"]+"_totalsize", totalSize, 5*60*time.Second)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	// 按照顺序排序
	res.Data = map[string]interface{}{
		"token":             token,
		"uploadedTotalSize": totalSize,
		"uploadedOffset":    uploadedOffset,
		"urls":              methods.GetResponseData(&fileConfigInfo),
	}
	res.Call(c)
}

// 如果配置文件存在，则token有效
// 如果配置文件不存在，则token无效
func (fc *ChunkUploadController) UploadChunk(c *gin.Context) {
	var res response.ResponseType
	res.Code = 200
	// log.Info("------UploadChunk------")
	fileConfigInfo := c.MustGet("fileConfigInfo").(*typings.TempFileConfigInfo)
	//

	// 获取数据库，检测文件是否上传成功过

	// log.Info("fileConfigInfo", fileConfigInfo)

	// 检查文件是否存在
	file, err := c.FormFile("files")
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}

	fileInfoMap := make(map[string]string)
	fileName, err := url.QueryUnescape(file.Filename)
	if err != nil {
		c.String(http.StatusOK, "error")
		return
	}
	err = json.Unmarshal([]byte(fileName), &fileInfoMap)
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}
	// log.Info("fileInfoMap", fileInfoMap["final"])

	totalSize, err := conf.Redisdb.Get("file_" + fileConfigInfo.FileInfo.Hash + "_totalsize")
	if err != nil {
		res.Error = err.Error()
		res.Code = 10001
		res.Call(c)
		return
	}

	// 当final等于no的同时size等于0，则不允许
	if fileInfoMap["final"] == "no" && file.Size == 0 {
		res.Code = 10016
		res.Call(c)
		return
	}
	// log.Info("final", fileInfoMap["final"])
	if fileInfoMap["final"] == "no" && file.Size != fileConfigInfo.ChunkSize {
		res.Code = 10018
		res.Call(c)
		return
	}
	err = c.SaveUploadedFile(file, fileConfigInfo.TempChuckFolderPath+fileInfoMap["offset"])
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}
	fileHash, err := methods.GetHash(fileConfigInfo.TempChuckFolderPath + fileInfoMap["offset"])
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}
	if fileHash != fileInfoMap["hash"] {
		res.Code = 10017
		res.Call(c)
		return
	}
	// log.Info("size", file.Size, fileConfigInfo.ChunkSize)

	// 当size到最后的时候
	totalSizeInt64, err := strconv.ParseInt(totalSize.String(), 10, 64)
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}
	// log.Info(totalSizeInt64+file.Size, fileConfigInfo.Size)
	// 已经全部传完
	if totalSizeInt64+file.Size == fileConfigInfo.FileInfo.Size {

		code, err := methods.MergeFiles(fileConfigInfo)
		if code == 200 {
			// 创建静态文件

			file := models.File{
				AppId:          fileConfigInfo.AppId,
				EncryptionName: fileConfigInfo.EncryptionName,
				FileName:       fileConfigInfo.Name,
				Path:           fileConfigInfo.Path,
				AvailableRange: models.FileAvailableRange{
					VisitCount:     fileConfigInfo.VisitCount,
					ExpirationTime: fileConfigInfo.ExpirationTime,
				},
				HashHistory: []models.HashHistory{},
				Hash:        fileConfigInfo.FileInfo.Hash,
			}
			saveFile, err := fileDbx.SaveFile(&file)
			log.Info(saveFile, err)
			if err != nil {
				res.Code = 10016
				res.Error = err.Error()
				res.Call(c)
				return
			}
			fileConfigInfo.EncryptionName = saveFile.EncryptionName
			res.Data = methods.GetResponseData(fileConfigInfo)
		}
		res.Code = code
		res.Error = err.Error()
		res.Code = 200
		res.Call(c)
		return
	} else {

		// 后续启用redis
		err = conf.Redisdb.Set("file_"+fileConfigInfo.FileInfo.Hash, c.MustGet("token").(string), 5*60*time.Second)
		if err != nil {
			res.Error = err.Error()
			res.Code = 10001
			res.Call(c)
			return
		}
		err = conf.Redisdb.Set("file_"+fileConfigInfo.FileInfo.Hash+"_totalsize", totalSizeInt64+file.Size, 5*60*time.Second)
		if err != nil {
			res.Error = err.Error()
			res.Code = 10001
			res.Call(c)
			return
		}
	}
	// if file.Size < fileConfigInfo.ChunkSize {
	// 	code, err, data := MergeFiles(fileConfigInfo)
	// 	res.Code = code
	// 	res.Error = err.Error()
	// 	res.Data = data
	// 	return
	// }

	res.Code = 200
	res.Call(c)
}
