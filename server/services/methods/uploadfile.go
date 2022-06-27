package methods

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/golang-jwt/jwt"
)

var (
	log = nlog.New()
)

func IsExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// func GetHash(file []byte) (string, error) {
// 	hasher := sha256.New()
// 	hasher.Write(file)
// 	return hex.EncodeToString(hasher.Sum(nil)), nil
// }

func GetHash(filePath string) (string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(filePath)
	hasher.Write(s)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

type jwtCustomClaims struct {
	// 追加自己需要的信息
	FileInfo typings.TempFileConfigInfo `json:"fileInfo"`
	jwt.StandardClaims
}

func GetToken(fileInfo typings.TempFileConfigInfo) (string, error) {
	claims := jwtCustomClaims{
		fileInfo,
		jwt.StandardClaims{
			ExpiresAt: int64(time.Now().Add(time.Hour * 24).Unix()),
			Issuer:    "saass",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(conf.FileTokenSign))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseToken(token string) (*typings.TempFileConfigInfo, error) {
	tokenData, err := jwt.ParseWithClaims(token, &jwtCustomClaims{}, func(tokenStr *jwt.Token) (interface{}, error) {
		if _, ok := tokenStr.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", tokenStr.Header["alg"])
		}
		return []byte(conf.FileTokenSign), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tokenData.Claims.(*jwtCustomClaims); ok && tokenData.Valid {
		if claims.FileInfo != (typings.TempFileConfigInfo{}) {
			return &claims.FileInfo, nil
		}
		return nil, err
	} else {
		return nil, err
	}
}

// 合并文件
func MergeFiles(fileConfigInfo *typings.TempFileConfigInfo) (code int64, err error) {

	// log.Info("meger start")

	if !IsExists(fileConfigInfo.StaticFolderPath) {
		os.MkdirAll(fileConfigInfo.StaticFolderPath, os.ModePerm)
	}

	filePath := fileConfigInfo.StaticFolderPath + fileConfigInfo.StaticFileName
	complateFile, err := os.Create(filePath)

	if err != nil {
		log.Info(err)
		code = 10016
		return
	}
	defer complateFile.Close()

	for i := int64(0); i <= fileConfigInfo.FileInfo.Size/fileConfigInfo.ChunkSize; i++ {

		fileBuffer, err := ioutil.ReadFile(fileConfigInfo.TempChuckFolderPath + strconv.FormatInt(i*fileConfigInfo.ChunkSize, 10))

		if err != nil {
			log.Info(err)
			code = 10016
			break
		}
		complateFile.Write(fileBuffer)
	}

	// log.Info("filePath", filePath)

	hash, err := GetHash(filePath)
	if err != nil {
		log.Info(err)
		code = 10016
		return
	}
	// log.Info(filePath)
	// log.Info(hash)
	// log.Info(fileConfigInfo.Hash)
	// log.Info("size", complateFile.Size())

	if fileConfigInfo.FileInfo.Hash != hash {
		code = 10016
		return
	}

	conf.Redisdb.Delete("file_" + fileConfigInfo.FileInfo.Hash)
	conf.Redisdb.Delete("file_" + fileConfigInfo.FileInfo.Hash + "_totalsize")

	// 全部删除所有临时文件
	os.RemoveAll(fileConfigInfo.TempFolderPath)
	// folderPath
	// 存储内容 hash

	code = 200
	err = errors.New("")

	return
}

func GetResponseData(fileConfigInfo *typings.TempFileConfigInfo) map[string]string {
	return map[string]string{
		"domainUrl":     conf.Config.StaticPathDomain,
		"encryptionUrl": "/s/" + fileConfigInfo.EncryptionName,
		"url":           "/s" + fileConfigInfo.Path + fileConfigInfo.Name + "?a=" + conf.AppList[fileConfigInfo.AppId].EncryptionId,
	}
}

// 获取上传进度
func GetUploadedOffset(fileConfigInfo *typings.TempFileConfigInfo) (int64, []int64) {
	totalSize := int64(0)
	uploadedOffset := []int64{}
	for i := int64(0); i <= fileConfigInfo.FileInfo.Size/fileConfigInfo.ChunkSize; i++ {
		fileBuffer, _ := ioutil.ReadFile(fileConfigInfo.TempChuckFolderPath + strconv.FormatInt(i*fileConfigInfo.ChunkSize, 10))
		totalSize += int64(len(fileBuffer))
		if int64(len(fileBuffer)) > 0 {
			uploadedOffset = append(uploadedOffset, i*fileConfigInfo.ChunkSize)
		}
	}
	return totalSize, uploadedOffset
}
