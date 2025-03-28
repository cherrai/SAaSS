package typings

import "go.mongodb.org/mongo-driver/bson/primitive"

type TempFileConfigInfo struct {
	AppId string
	Name  string
	// 如果文件一样，则加密地址永远也是一样的
	ShortId string
	// 文件存储路径
	RootPath         string
	ParentFolderPath string
	ParentFolderId   primitive.ObjectID

	// StaticFolderPath string
	// StaticFileName   string
	// 临时文件夹路径
	TempFolderPath      string
	TempChuckFolderPath string
	// Type                string
	ChunkSize      int64
	CreateTime     int64
	ExpirationTime int64
	VisitCount     int64
	Password       string
	FileInfo       FileInfo
	UserId         string
	UploadUserId   string
	AllowShare     int64
	ShareUsers     []string
	// Parameter      Parameter
}

// type Parameter struct {
// 	// x-saass-process=image/resize,160,70
// 	Parameter string
// 	Hash      string
// }

type FileInfo struct {
	Name         string
	Size         int64
	Type         string
	Suffix       string
	LastModified int64
	Hash         string
}
