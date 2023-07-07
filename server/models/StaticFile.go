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

// // 如果源文件hash一样
// // 参数也是一样，那就是重复上传
// type Process struct {
// 	// x-saass-process=image/resize,160,70
// 	Parameter string `bson:"parameter" json:"parameter,omitempty"`
// 	// 仅仅存储Hash即可
// 	ProcessInfo ProcessInfo `bson:"processInfo" json:"processInfo,omitempty"`
// }
// type ProcessInfo struct {
// 	Hash string `bson:"hash" json:"hash,omitempty"`
// }
type FileInfo struct {
	Name         string `bson:"name" json:"name,omitempty"`
	Size         int64  `bson:"size" json:"size,omitempty"`
	Type         string `bson:"type" json:"type,omitempty"`
	Suffix       string `bson:"suffix" json:"suffix,omitempty"`
	LastModified int64  `bson:"lastModified" json:"lastModified,omitempty"`
	Hash         string `bson:"hash" json:"hash,omitempty"`
	Width        int64  `bson:"width" json:"width,omitempty"`
	Height       int64  `bson:"height" json:"height,omitempty"`
	// Process      Process `bson:"process" json:"process,omitempty"`
}

type StaticFile struct {
	// 加密IDName
	Id primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	// 文件名
	FileName string `bson:"fileName" json:"fileName,omitempty"`
	// 文件夹路径
	Path string `bson:"path" json:"path,omitempty"`

	FileInfo FileInfo `bson:"fileInfo" json:"fileInfo,omitempty"`

	// DeleteStatus:
	// 1 normal
	// 0 not accessible
	// -1 delete
	// -2 file_delete
	Status int `bson:"status" json:"status,omitempty"`
	// CreateTime Unix timestamp
	CreateTime int64 `bson:"createTime" json:"createTime,omitempty"`
	// UpdateTIme Unix timestamp
	UpdateTIme int64 `bson:"updateTIme" json:"updateTIme,omitempty"`
	// DeleteTime Unix timestamp
	DeleteTime int64 `bson:"deleteTime" json:"deleteTime,omitempty"`
}

func (m *StaticFile) GetCollectionName() string {
	return "StaticFiles"
}

func (m *StaticFile) Default() error {
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
	if m.UpdateTIme == 0 {
		m.UpdateTIme = unixTimeStamp
	}
	if m.DeleteTime == 0 {
		m.DeleteTime = -1
	}

	if err := m.Validate(); err != nil {
		return errors.New(m.GetCollectionName() + " Validate: " + err.Error())
	}
	return nil
}

func (m *StaticFile) GetCollection() *mongo.Collection {
	return mongodb.GetCollection(conf.Config.Mongodb.Currentdb.Name, m.GetCollectionName())
}

func (m *StaticFile) Validate() error {
	errStr := ""
	if m.FileInfo != (FileInfo{}) {
		err := validation.ValidateStruct(
			&m.FileInfo,
			validation.Parameter(&m.FileInfo.Name, validation.Required()),
			validation.Parameter(&m.FileInfo.Size, validation.Required(), validation.GreaterEqual(1)),
			validation.Parameter(&m.FileInfo.Type, validation.Required()),
			// validation.Parameter(&m.FileInfo.Suffix, validation.Required()),
			validation.Parameter(&m.FileInfo.LastModified, validation.Required()),
			validation.Parameter(&m.FileInfo.Hash, validation.Required()),
		)
		if err != nil {
			errStr += err.Error()
		}
	}
	err := validation.ValidateStruct(
		m,
		validation.Parameter(&m.FileName, validation.Required()),
		validation.Parameter(&m.Path, validation.Required()),
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
