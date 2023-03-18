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

type AvailableRangeShareUsers struct {
	// 如果有个uid是"AllUser",则是所有人都不可见
	// 非所有人则需要提供一个携带uid的字符串，每次都从SAaSS获取，有效时间5分钟
	Uid        string `bson:"uid" json:"uid"`
	CreateTime int64  `bson:"createTime" json:"createTime"`
}

type FileAvailableRange struct {
	// -1 就是不限制 非-1则就是要和隔壁的比较大小
	VisitCount int64 `bson:"visitCount" json:"visitCount,omitempty"`
	// ExpirationTime Unix timestamp
	ExpirationTime int64  `bson:"expirationTime" json:"expirationTime,omitempty"`
	Password       string `bson:"password" json:"password,omitempty"`
	// 创建人
	AuthorId string `bson:"authorId" json:"authorId,omitempty"`
	// 是否允许共享
	// 2 允许编辑 预留
	// 1 允许下载
	// -1 私有不允许外部下载（私有的话，也需要检测UID
	AllowShare int64                       `bson:"allowShare" json:"allowShare,omitempty"`
	ShareUsers []*AvailableRangeShareUsers `bson:"shareUsers" json:"shareUsers,omitempty"`
}
type FileUsage struct {
	VisitCount int64 `bson:"visitCount" json:"visitCount,omitempty"`
}
type HashHistory struct {
	Hash string `bson:"hash" json:"hash,omitempty"`
}

type File struct {
	// 加密IDName
	Id    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	AppId string             `bson:"appId" json:"appId,omitempty"`
	// 加密文件名
	ShortId string `bson:"shortId" json:"shortId,omitempty"`
	// 云盘文件名
	FileName string `bson:"fileName" json:"fileName,omitempty"`
	// 云盘存储与访问路径
	// Path string `bson:"path" json:"path,omitempty"`
	ParentFolderId primitive.ObjectID `bson:"parentFolderId" json:"parentFolderId,omitempty"`
	// Hash
	Hash string `bson:"hash" json:"hash,omitempty"`
	// Label
	Label string `bson:"label" json:"label,omitempty"`
	// replace update 历史记录
	HashHistory []*HashHistory `bson:"hashHistory" json:"hashHistory,omitempty"`

	AvailableRange FileAvailableRange `bson:"availableRange" json:"availableRange,omitempty"`
	Usage          FileUsage          `bson:"usage" json:"usage,omitempty"`

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
	// // DeadlineInRecycleBin Unix timestamp
	// DeadlineInRecycleBin int64 `bson:"deadlineInRecycleBin" json:"deadlineInRecycleBin,omitempty"`
	// LastDownloadTime Unix timestamp
	LastDownloadTime int64 `bson:"lastDownloadTime" json:"lastDownloadTime,omitempty"`
}

func (m *File) GetCollectionName() string {
	return "Files"
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
	if m.LastUpdateTime == 0 {
		m.LastUpdateTime = unixTimeStamp
	}
	if m.DeleteTime == 0 {
		m.DeleteTime = -1
	}
	if m.HashHistory == nil {
		m.HashHistory = []*HashHistory{}
	}
	// if m.DeadlineInRecycleBin == 0 {
	// 	m.DeadlineInRecycleBin = -1
	// }
	if m.LastDownloadTime == 0 {
		m.LastDownloadTime = -1
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
	err := validation.ValidateStruct(
		m,
		validation.Parameter(&m.AppId, validation.Required()),
		validation.Parameter(&m.ShortId, validation.Required()),
		validation.Parameter(&m.FileName, validation.Required()),
		// validation.Parameter(&m.ParentFolderId, validation.Required()),
		validation.Parameter(&m.Hash, validation.Required()),
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
