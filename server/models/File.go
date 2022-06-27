package models

import (
	"errors"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	mongodb "github.com/cherrai/SAaSS/db/mongo"

	"github.com/cherrai/nyanyago-utils/validation"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileInfo struct {
	Name         string `bson:"name" json:"name,omitempty"`
	Size         int64  `bson:"size" json:"size,omitempty"`
	Type         string `bson:"type" json:"type,omitempty"`
	Suffix       string `bson:"suffix" json:"suffix,omitempty"`
	LastModified int64  `bson:"lastModified" json:"lastModified,omitempty"`
	Hash         string `bson:"hash" json:"hash,omitempty"`
	Width        int64  `bson:"width" json:"width,omitempty"`
	Height       int64  `bson:"height" json:"height,omitempty"`
}

type FileAvailableRange struct {
	// -1 就是不限制 非-1则就是要和隔壁的比较大小
	VisitCount int64 `bson:"visitCount" json:"visitCount,omitempty"`
	// ExpirationTime Unix timestamp
	ExpirationTime int64 `bson:"expirationTime" json:"expirationTime,omitempty"`
}
type FileUsage struct {
	VisitCount int64 `bson:"visitCount" json:"visitCount,omitempty"`
}

type File struct {
	// 加密IDName
	Id    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	AppId string             `bson:"appId" json:"appId,omitempty"`
	// 加密文件名
	EncryptionName string `bson:"encryptionName" json:"encryptionName,omitempty"`
	// 云盘文件名
	FileName string `bson:"fileName" json:"fileName,omitempty"`
	// 云盘存储与访问路径
	Path string `bson:"path" json:"path,omitempty"`
	// 文件类型
	Type string `bson:"type" json:"type,omitempty"`
	// 静态存储文件夹
	// 本地文件实际存储时间为3个月
	StaticFolderPath string `bson:"staticFolderPath" json:"staticFolderPath,omitempty"`
	// 静态存储名
	StaticFileName string             `bson:"staticFileName" json:"staticFileName,omitempty"`
	FileInfo       FileInfo           `bson:"fileInfo" json:"fileInfo,omitempty"`
	AvailableRange FileAvailableRange `bson:"availableRange" json:"availableRange,omitempty"`
	Usage          FileUsage          `bson:"usage" json:"usage,omitempty"`
	// DeleteStatus:
	// 1 normal
	// 0 not accessible
	// -1 delete
	// -2 file_delete
	Status int `bson:"status" json:"status,omitempty"`
	// CreateTime Unix timestamp
	CreateTime int64 `bson:"createTime" json:"createTime,omitempty"`
	// DeleteTime Unix timestamp
	DeleteTime int64 `bson:"deleteTime" json:"deleteTime,omitempty"`
}

func (m *File) GetCollectionName() string {
	return "File"
}

func (m *File) Default() error {
	if m.Id == primitive.NilObjectID {
		m.Id = primitive.NewObjectID()
	}
	if m.Status == 0 {
		m.Status = 1
	}
	unixTimeStamp := time.Now().Unix()
	if m.CreateTime == 0 {
		m.CreateTime = unixTimeStamp
	}
	if m.DeleteTime == 0 {
		m.DeleteTime = -1
	}
	if m.AvailableRange.VisitCount <= 0 {
		m.AvailableRange.VisitCount = -1
	}
	if m.AvailableRange.ExpirationTime == 0 {
		m.AvailableRange.ExpirationTime = -1
	}

	if err := m.Validate(); err != nil {
		return errors.New(m.GetCollectionName() + " Validate: " + err.Error())
	}
	return nil
}

func (m *File) GetCollection() *mongo.Collection {
	return mongodb.GetCollection(conf.Config.Mongodb.Currentdb.Name, m.GetCollectionName())
}

func (m *File) Validate() error {
	errStr := ""
	if m.FileInfo != (FileInfo{}) {
		err := validation.ValidateStruct(
			&m.FileInfo,
			validation.Parameter(&m.FileInfo.Name, validation.Required()),
			validation.Parameter(&m.FileInfo.Size, validation.Required(), validation.GreaterEqual(1)),
			validation.Parameter(&m.FileInfo.Type, validation.Required()),
			validation.Parameter(&m.FileInfo.Suffix, validation.Required()),
			validation.Parameter(&m.FileInfo.LastModified, validation.Required()),
			validation.Parameter(&m.FileInfo.Hash, validation.Required()),
		)
		if err != nil {
			errStr += err.Error()
		}
	}
	err := validation.ValidateStruct(
		m,
		validation.Parameter(&m.AppId, validation.Required()),
		validation.Parameter(&m.EncryptionName, validation.Required()),
		validation.Parameter(&m.FileName, validation.Required()),
		validation.Parameter(&m.Path, validation.Required()),
		validation.Parameter(&m.Type, validation.Required()),
		validation.Parameter(&m.StaticFolderPath, validation.Required()),
		validation.Parameter(&m.StaticFileName, validation.Required()),
		validation.Parameter(&m.Status, validation.Enum([]int{1, -1, -2})),
		validation.Parameter(&m.CreateTime, validation.Required()),
	)
	if err != nil {
		errStr += err.Error()
	}
	if errStr == "" {
		return nil
	}
	return errors.New(errStr)
}
