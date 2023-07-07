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
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nint"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/cherrai/nyanyago-utils/validation"
	"github.com/gin-gonic/gin"
)

var (
	log = conf.Log
)

var fileDbx = new(dbxV1.FileDbx)

type ChunkUploadController struct {
}

// 如果接着上次的传，则需要获取上次的上传进度
func (fc *ChunkUploadController) CreateChunk(c *gin.Context) {
	// log.Info("------CreateChunk------")
	var res response.ResponseType
	res.Code = 200

	appId := c.GetString("appId")
	userId := c.GetString("userId")
	rootPath := nstrings.StringOr(c.PostForm("rootPath"), "/")

	ati, exists := c.Get("appTokenInfo")
	if exists {
		t := ati.(*typings.AppTokenInfo)
		rootPath = t.RootPath
	}

	allowShare := nint.ToInt64(c.PostForm("allowShare"))
	shareUsers := []string{}

	// log.Info("ShareUsers", shareUsers)
	// log.Info("ShareUsers", c.PostForm("shareUsers"))
	// log.Info("ShareUsers", c.PostFormMap("shareUsers"))

	for _, v := range c.PostFormMap("shareUsers") {
		shareUsers = append(shareUsers, v)
	}
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
	tempFolderPath := "./static/chuck/" + fileInfo["hash"] + "/" + c.PostForm("chunkSize") + "/"

	fileNameWithSuffix := path.Base(fileInfo["name"])
	fileType := path.Ext(fileNameWithSuffix)
	fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)

	parentFolderPath := path.Join(rootPath, c.PostForm("path"))
	parentFolderId, err := folderDbx.GetParentFolderId(appId, parentFolderPath, true, userId)
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}

	shortId, err := fileDbx.GetShortId(9)
	if err != nil {
		res.Code = 10016
		res.Error = err.Error()
		res.Call(c)
		return
	}

	fileConfigInfo := typings.TempFileConfigInfo{
		AppId:   appId,
		Name:    fileName,
		ShortId: shortId,
		// EncryptionName:   strings.ToLower(cipher.MD5(fileInfo["hash"] + appId + fileInfo["size"] + nstrings.ToString(time.Now().Unix()))),
		RootPath:         rootPath,
		ParentFolderPath: parentFolderPath,
		ParentFolderId:   parentFolderId,
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
		Password:       c.PostForm("password"),
		FileInfo: typings.FileInfo{
			Name:         fileNameOnly,
			Size:         nint.ToInt64(fileInfo["size"]),
			Type:         fileInfo["type"],
			Suffix:       fileType,
			LastModified: nint.ToInt64(fileInfo["lastModified"]),
			Hash:         fileInfo["hash"],
		},
		UserId:     userId,
		AllowShare: allowShare,
		ShareUsers: shareUsers,
	}

	if err = validation.ValidateStruct(
		&fileConfigInfo,
		validation.Parameter(&fileConfigInfo.Name, validation.Required()),
		validation.Parameter(&fileConfigInfo.ParentFolderPath, validation.Required()),
		validation.Parameter(&fileConfigInfo.TempFolderPath, validation.Required()),
		validation.Parameter(&fileConfigInfo.TempChuckFolderPath, validation.Required()),
		validation.Parameter(&fileConfigInfo.ChunkSize, validation.Type("int64"), validation.Required(),
			validation.Enum([]int64{
				16 * 1024,
				32 * 1024,
				64 * 1024,
				128 * 1024,
				256 * 1024,
				512 * 1024,
				1024 * 1024,
				2 * 1024 * 1024,
				5 * 1024 * 1024,
				10 * 1024 * 1024,
				20 * 1024 * 1024,
				30 * 1024 * 1024,
				50 * 1024 * 1024,
			}), validation.GreaterEqual(1)),
		validation.Parameter(&fileConfigInfo.UserId, validation.Type("string"), validation.Required()),
		validation.Parameter(&fileConfigInfo.AllowShare, validation.Type("int64"), validation.Enum([]int64{2, 1, -1}), validation.Required()),
		// validation.Parameter(&fileConfigInfo.ShareUsers, validation.Required()),
		validation.Parameter(&fileConfigInfo.RootPath, validation.Type("string"), validation.Required()),
		// validation.Parameter(&fileConfigInfo.Type, validation.Type("string"), validation.Required(), validation.Enum([]string{"Image", "Video", "Audio", "Text", "File"})),
	); err != nil {
		res.Error = err.Error()
		res.Code = 10002
		res.Call(c)
		return
	}
	err = validation.ValidateStruct(
		&fileConfigInfo.FileInfo,
		validation.Parameter(&fileConfigInfo.FileInfo.Name, validation.Required()),
		validation.Parameter(&fileConfigInfo.FileInfo.Size, validation.Type("int64"), validation.Required(), validation.GreaterEqual(1)),
		// validation.Parameter(&fileConfigInfo.FileInfo.Type, validation.Required()),
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
	file, err := fileDbx.GetFileWithFileInfo(fileConfigInfo.AppId, fileConfigInfo.ParentFolderPath, fileConfigInfo.Name, userId)
	log.Info("file", file)
	if err != nil {
		res.Error = err.Error()
		res.Code = 10019
		res.Call(c)
		return
	}

	// 内容存在
	staticFilesIsExist := false

	log.Info(fileConfigInfo.FileInfo.Hash)
	sf, err := fileDbx.GetStaticFileWithHash(fileConfigInfo.FileInfo.Hash)
	log.Info(sf, err)
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
	log.Info("staticFilesIsExist", staticFilesIsExist)
	if staticFilesIsExist {
		// 内容存在则更新

		if file != nil && file.Hash != "" {
			// 更新内容到最新状态
			if file.Hash != fileConfigInfo.FileInfo.Hash {
				file.HashHistory = append(file.HashHistory, &models.HashHistory{
					Hash: file.Hash,
				})
				file.Hash = fileConfigInfo.FileInfo.Hash
			}
			file.Status = 1
			file.DeleteTime = -1
			file.AvailableRange.VisitCount = fileConfigInfo.VisitCount
			file.AvailableRange.Password = fileConfigInfo.Password
			file.AvailableRange.AllowShare = fileConfigInfo.AllowShare
			file.AvailableRange.ShareUsers = []*models.AvailableRangeShareUsers{}
			for _, v := range fileConfigInfo.ShareUsers {
				file.AvailableRange.ShareUsers = append(file.AvailableRange.ShareUsers, &models.AvailableRangeShareUsers{
					Uid:        v,
					CreateTime: time.Now().Unix(),
				})
			}
			file.AvailableRange.ExpirationTime = fileConfigInfo.ExpirationTime
			fileConfigInfo.ShortId = file.ShortId

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

			_, err := fc.SaveFile(&fileConfigInfo)

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
	} else {
		if file != nil && file.Hash != "" && fileConfigInfo.ParentFolderId == file.ParentFolderId && fileConfigInfo.Name == file.FileName {
			fileConfigInfo.ShortId = file.ShortId
		}
	}

	// 创建文件信息的临时配置文件
	if !nfile.IsExists(fileConfigInfo.TempChuckFolderPath) {
		os.MkdirAll(fileConfigInfo.TempChuckFolderPath, os.ModePerm)
	}
	// 保存临时配置文件
	serverConfig, serverConfigErr := os.Create(fileConfigInfo.TempFolderPath + "/info.json")
	if serverConfigErr != nil {
		res.Error = serverConfigErr.Error()
		res.Code = 10001
		res.Call(c)
		return
	}
	defer serverConfig.Close()

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
	token, err := methods.GetToken(&fileConfigInfo)
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
	log.Info("cccccccccccc", totalSize, uploadedOffset)
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
	log.Info(res.Data)
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
	log.Info("final", fileInfoMap["final"], file.Size, fileConfigInfo.ChunkSize)
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
	log.Info(totalSizeInt64, file.Size, totalSizeInt64+file.Size, fileConfigInfo.FileInfo.Size)
	log.Info(totalSizeInt64 == file.Size || totalSizeInt64+file.Size == fileConfigInfo.FileInfo.Size)
	// 已经全部传完
	// totalSizeInt64 == file.Size ||
	if totalSizeInt64+file.Size == fileConfigInfo.FileInfo.Size {

		code, err := methods.MergeFiles(fileConfigInfo)
		log.Info("MergeFiles", code, err != nil)
		if err != nil {
			res.Code = 10016
			res.Error = err.Error()
			res.Call(c)
			return
		}
		if code == 200 {
			// 创建静态文件

			saveFile, err := fc.SaveFile(fileConfigInfo)

			log.Info("saveFile", fileConfigInfo.Password, saveFile, err)
			if err != nil {
				res.Code = 10016
				res.Error = err.Error()
				res.Call(c)
				return
			}
			fileConfigInfo.ShortId = saveFile.ShortId
			res.Data = methods.GetResponseData(fileConfigInfo)
		}
		res.Code = code
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
	// res.Data = methods.GetResponseData(fileConfigInfo)
	res.Call(c)
}

func (fc *ChunkUploadController) SaveFile(fConfig *typings.TempFileConfigInfo) (*models.File, error) {
	su := []*models.AvailableRangeShareUsers{}
	for _, v := range fConfig.ShareUsers {
		su = append(su, &models.AvailableRangeShareUsers{
			Uid:        v,
			CreateTime: time.Now().Unix(),
		})
	}

	log.Info("su", su)

	file := models.File{
		AppId:          fConfig.AppId,
		ShortId:        fConfig.ShortId,
		FileName:       fConfig.Name,
		ParentFolderId: fConfig.ParentFolderId,
		Status:         1,
		AvailableRange: models.FileAvailableRange{
			VisitCount:     fConfig.VisitCount,
			ExpirationTime: fConfig.ExpirationTime,
			Password:       fConfig.Password,
			AuthorId:       fConfig.UserId,
			AllowShare:     fConfig.AllowShare,
			ShareUsers:     su,
		},
		Hash: fConfig.FileInfo.Hash,
	}
	_, err := fileDbx.SaveFile(&file)
	log.Info("SaveFile", err)
	if err != nil {
		return nil, err
	}
	return &file, nil
}
