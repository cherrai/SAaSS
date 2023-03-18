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

type FolderAvailableRange struct {
	Password string `bson:"password" json:"password,omitempty"`
	// 创建人
	AuthorId string `bson:"authorId" json:"authorId,omitempty"`
	// 是否允许共享
	// 2 允许编辑 预留
	// 1 允许下载
	// -1 私有不允许外部下载（私有的话，也需要检测UID
	AllowShare int64                       `bson:"allowShare" json:"allowShare,omitempty"`
	ShareUsers []*AvailableRangeShareUsers `bson:"shareUsers" json:"shareUsers,omitempty"`
}
type FolderUsage struct {
}

type Folder struct {
	Id    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	AppId string             `bson:"appId" json:"appId,omitempty"`
	// 加密文件名
	ShortId string `bson:"shortId" json:"shortId,omitempty"`
	// 文件夹名
	FolderName string `bson:"folderName" json:"folderName,omitempty"`
	// 父级文件夹路径Id
	ParentFolderId primitive.ObjectID `bson:"parentFolderId" json:"parentFolderId,omitempty"`
	// // 父级文件夹路径
	// ParentPath string `bson:"parentPath" json:"parentPath,omitempty"`

	AvailableRange FolderAvailableRange `bson:"availableRange" json:"availableRange,omitempty"`
	Usage          FolderUsage          `bson:"usage" json:"usage,omitempty"`

	// DeleteStatus:
	// 1 normal
	// 0 not accessible
	// -1 RecycleBin
	// -2 delete
	Status int64 `bson:"status" json:"status,omitempty"`
	// CreateTime Unix timestamp
	CreateTime int64 `bson:"createTime" json:"createTime,omitempty"`
	// UpdateTime Unix timestamp
	LastUpdateTime int64 `bson:"lastUpdateTime" json:"lastUpdateTime,omitempty"`
	// DeleteTime Unix timestamp
	DeleteTime int64 `bson:"deleteTime" json:"deleteTime,omitempty"`
}

func (m *Folder) GetCollectionName() string {
	return "Folders"
}

func (m *Folder) Default() error {
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
	if m.LastUpdateTime == 0 {
		m.LastUpdateTime = unixTimeStamp
	}
	if m.DeleteTime == 0 {
		m.DeleteTime = -1
	}
	if err := m.Validate(); err != nil {
		return errors.New(m.GetCollectionName() + " Validate: " + err.Error())
	}
	return nil
}

func (m *Folder) GetCollection() *mongo.Collection {
	return mongodb.GetCollection(conf.Config.Mongodb.Currentdb.Name, m.GetCollectionName())
}

func (m *Folder) Validate() error {
	errStr := ""
	err := validation.ValidateStruct(
		m,
		validation.Parameter(&m.AppId, validation.Required()),
		validation.Parameter(&m.ShortId, validation.Required()),
		validation.Parameter(&m.FolderName, validation.Required()),
		validation.Parameter(&m.Status, validation.Enum([]int64{1, -1, -2})),
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
