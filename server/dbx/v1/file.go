package dbxV1

import (
	"context"
	"errors"
	"path"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nshortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	log       = conf.Log
	fileDbx   = FileDbx{}
	folderDbx = FolderDbx{}
)

type FileDbx struct {
}

func (fd *FileDbx) GetAllTempFile(staticFileName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"staticFileName": staticFileName,
					},
				},
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) FileNotAccessible(id primitive.ObjectID) error {
	file := new(models.File)
	_, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"_id": id,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"status": 0,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	return nil
}

func (fd *FileDbx) ExpiredFile(id primitive.ObjectID) error {
	file := new(models.File)
	_, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"_id": id,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"status":     -1,
				"deleteTime": time.Now().Unix(),
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	return nil
}

func (fd *FileDbx) VisitFile(id primitive.ObjectID) error {
	file := new(models.File)
	_, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"_id": id,
				},
			},
		}, bson.M{
			"$inc": bson.M{
				"usage.visitCount": 1,
			},
			"$set": bson.M{
				"lastDownloadTime": time.Now().Unix(),
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	return nil
}

func (fd *FileDbx) RenameFile(appId, path, oldFileName, newFileName string, authorId string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	getFile, err := fd.GetFileWithFileInfo(appId, path, newFileName, authorId)
	if err != nil {
		return err
	}
	if getFile != nil {
		return errors.New("the filename is duplicated")
	}
	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"fileName": oldFileName,
				},
				{
					"status": 1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"lastlastUpdateTime": time.Now().Unix(),
				"fileName":           newFileName,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	if updateResult.ModifiedCount == 0 {
		return errors.New("update fail")
	}
	return nil
}

func (fd *FileDbx) MoveFilesToTrash(appId, path string, fileNames []string, authorId string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"fileName": bson.M{
						"$in": fileNames,
					},
				},
				{
					"status": 1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"deleteTime": time.Now().Unix(),
				"status":     -1,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	if updateResult.ModifiedCount == 0 {
		return errors.New("delete fail")
	}
	return nil
}

func (fd *FileDbx) Restore(appId, path string, fileNames []string, authorId string) error {
	err := fd.DeleteFiles(appId, path, fileNames, authorId, []int64{1, 0})
	if err != nil {
		return err
	}
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"fileName": bson.M{
						"$in": fileNames,
					},
				},
				{
					"status": -1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"deleteTime": -1,
				"status":     1,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	if updateResult.ModifiedCount == 0 {
		return errors.New("restore fail")
	}
	return nil
}

func (fd *FileDbx) DeleteFiles(appId, path string, fileNames []string, authorId string, status []int64) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	params := bson.M{
		"$and": []bson.M{
			{
				"appId": appId,
			},
			{
				"parentFolderId": parentFolderId,
			},
			{
				"status": bson.M{
					"$in": status,
				},
			},
			{
				"fileName": bson.M{
					"$in": fileNames,
				},
			},
		},
	}

	_, err = file.GetCollection().DeleteMany(context.TODO(), params)
	if err != nil {
		return err
	}
	return nil
}

func (fd *FileDbx) SetFileSharing(appId, authorId, path string, fileNames []string, status int64) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"fileName": bson.M{
						"$in": fileNames,
					},
				},
				{
					"status": 1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"lastlastUpdateTime":        time.Now().Unix(),
				"availableRange.allowShare": status,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	if updateResult.ModifiedCount == 0 {
		return errors.New("update fail")
	}
	return nil
}

func (fd *FileDbx) SetFilePassword(appId, authorId, path, fileName, password string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.File)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"fileName": fileName,
				},
				{
					"status": 1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"lastUpdateTime":          time.Now().Unix(),
				"availableRange.password": password,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	if updateResult.ModifiedCount == 0 {
		return errors.New("update fail")
	}
	return nil
}

func (fd *FileDbx) CopyFile(appId, authorId, path string, fileNames []string, newPath string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	newParentFolderId, err := folderDbx.GetParentFolderId(appId, newPath, false, authorId)
	if err != nil {
		return err
	}

	files, err := fd.GetFileLisByParentFolderIdOrFileNames(appId, parentFolderId, fileNames)
	if err != nil {
		return err
	}

	for _, v := range files.List {
		v.Id = primitive.NilObjectID
		shortId, err := fileDbx.GetShortId(9)
		if err != nil {
			return err
		}
		v.ShortId = shortId
		v.ParentFolderId = newParentFolderId
		_, err = fd.SaveFile(v)
		// log.Error(err)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fd *FileDbx) MoveFile(appId, authorId, path string, fileNames []string, newPath string) error {
	err := fd.CopyFile(appId, authorId, path, fileNames, newPath)
	if err != nil {
		return err
	}

	return fd.DeleteFiles(appId, path, fileNames, authorId, []int64{1, 0})
}

type GetFileListByPathType struct {
	List  ([]*models.File)
	Total []struct {
		Count int64
	}
}

func (fd *FileDbx) GetFileListByPath(appId, path string, authorId string) (*GetFileListByPathType, error) {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)

	if err != nil {
		return nil, err
	}
	return fd.GetFileLisByParentFolderId(appId, parentFolderId)
}

func (fd *FileDbx) GetFileLisByParentFolderId(appId string, parentFolderId primitive.ObjectID) (*GetFileListByPathType, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"parentFolderId": parentFolderId,
						// "path": bson.M{
						// 	"$regex": primitive.Regex{
						// 		Pattern: "^" + path + ".*",
						// 		Options: "i",
						// 	},
						// },
					},
					{
						"status": 1,
					},
				},
			},
		},
		{
			"$facet": bson.M{
				"list": []bson.M{
					{
						"$sort": bson.M{
							"lastUpdateTime": -1,
							"createTime":     -1,
						},
					},
					// {"$skip": pageSize * (pageNum - 1)}, {"$limit": pageSize},
				},
				"total": []bson.M{
					{"$count": "count"},
				},
			},
		},
	}

	var results []*GetFileListByPathType
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &GetFileListByPathType{}, nil
	}
	return results[0], nil
}

func (fd *FileDbx) GetFileLisByParentFolderIdList(appId string, parentFolderIdList []primitive.ObjectID, pageNum, pageSize int64, status []int64) ([]*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"parentFolderId": bson.M{
							"$in": parentFolderIdList,
						},
					},
					{
						"status": bson.M{
							"$in": status,
						},
					},
				},
			},
		},
		{
			"$sort": bson.M{
				"lastUpdateTime": -1,
				"createTime":     -1,
			},
		},
		{
			"$skip": pageSize * (pageNum - 1),
		},
		{
			"$limit": pageSize,
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return results, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return results, err
	}
	if len(results) == 0 {
		return results, nil
	}
	return results, nil
}

func (fd *FileDbx) GetFileLisByAuthorId(
	appId string,
	authorId string,
	pageNum,
	pageSize int64,
	status []int64) ([]*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"availableRange.authorId": authorId,
					},
					{
						"status": bson.M{
							"$in": status,
						},
					},
				},
			},
		},
		{
			"$sort": bson.M{
				"lastUpdateTime": -1,
				"createTime":     -1,
			},
		},
		{
			"$skip": pageSize * (pageNum - 1),
		},
		{
			"$limit": pageSize,
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return results, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return results, err
	}
	if len(results) == 0 {
		return results, nil
	}
	return results, nil
}

func (fd *FileDbx) GetFileLisByParentFolderIdOrFileNames(appId string, parentFolderId primitive.ObjectID, fileNames []string) (*GetFileListByPathType, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"parentFolderId": parentFolderId,
					},
					{
						"fileName": bson.M{
							"$in": fileNames,
						},
					},
					{
						"status": 1,
					},
				},
			},
		},
		{
			"$facet": bson.M{
				"list": []bson.M{
					{
						"$sort": bson.M{
							"lastUpdateTime": -1,
							"createTime":     -1,
						},
					},
					// {"$skip": pageSize * (pageNum - 1)}, {"$limit": pageSize},
				},
				"total": []bson.M{
					{"$count": "count"},
				},
			},
		},
	}

	var results []*GetFileListByPathType
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return &GetFileListByPathType{}, nil
	}
	return results[0], nil
}

func (u *FileDbx) GetShortId(digits int) (string, error) {
	str := nshortid.GetSpecifiedRandomString("HIJKLMN", 1) + nshortid.GetShortId(digits)

	// 检测
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"shortId": str,
					},
				},
			},
		},
	}

	// log.Info("GetFileWithEncryptionName encryptionName", params, encryptionName)

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return "", err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return "", err
	}
	if len(results) == 0 {
		return str, nil
	}

	return u.GetShortId(digits)
}

func (fd *FileDbx) GetFileWithShortId(shortId string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"shortId": shortId,
					},
					{
						"status": bson.M{
							"$in": []int64{1},
						},
					},
				},
			},
		},
	}

	// log.Info("GetFileWithEncryptionName encryptionName", params, encryptionName)

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		// log.Info(results, encryptionName, params)
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) GetFileWithPath(staticFileName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"staticFileName": staticFileName,
					},
					{
						"status": 1,
					},
				},
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) GetFileWithStaticFileName(staticFileName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"staticFileName": staticFileName,
					},
					{
						"status": 1,
					},
				},
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) GetFileWithFileInfo(appId string, path string, fileName string, authorId string) (*models.File, error) {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return nil, err
	}
	return fd.GetFileWithParentFolderId(appId, parentFolderId, fileName)
}

func (fd *FileDbx) GetFileWithParentFolderId(appId string, parentFolderId primitive.ObjectID, fileName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"parentFolderId": parentFolderId,
					},
					{
						"fileName": fileName,
					},
					{
						"status": 1,
					},
				},
				// and groupId
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) GetStaticFileWithHash(hash string) (*models.StaticFile, error) {
	staticFile := new(models.StaticFile)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"fileInfo.hash": hash,
					},
					{
						"status": 1,
					},
				},
			},
		},
	}

	var results []*models.StaticFile
	opts, err := staticFile.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (fd *FileDbx) GetStaticFileListWithHash(hashList []string) ([]*models.StaticFile, error) {
	staticFile := new(models.StaticFile)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"fileInfo.hash": bson.M{
							"$in": hashList,
						},
					},
					{
						"status": 1,
					},
				},
			},
		},
	}

	var results []*models.StaticFile
	opts, err := staticFile.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (fd *FileDbx) GetStaticFileWithPath(path, fileName string) (*models.StaticFile, error) {
	staticFile := new(models.StaticFile)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"path": path,
					},
					{
						"fileName": fileName,
					},
				},
			},
		},
	}

	var results []*models.StaticFile
	opts, err := staticFile.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (fd *FileDbx) UpdateFile(file *models.File) (*mongo.UpdateResult, error) {
	result, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"_id": file.Id,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"hash":           file.Hash,
				"status":         file.Status,
				"lastUpdateTime": time.Now().Unix(),
				"deleteTime":     file.DeleteTime,
				"availableRange": file.AvailableRange,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (fd *FileDbx) SaveFile(file *models.File) (*models.File, error) {
	// 先检测状态正常的有没有
	getFile, err := fd.GetFileWithParentFolderId(file.AppId, file.ParentFolderId, file.FileName)
	// log.Info(getFile, err)
	if err != nil {
		return nil, err
	}
	if getFile != nil {
		getFile.Status = 1
		getFile.DeleteTime = -1
		getFile.AvailableRange.VisitCount = file.AvailableRange.VisitCount
		getFile.AvailableRange.ExpirationTime = file.AvailableRange.ExpirationTime
		getFile.AvailableRange.Password = file.AvailableRange.Password
		getFile.AvailableRange.AllowShare = file.AvailableRange.AllowShare
		getFile.AvailableRange.ShareUsers = file.AvailableRange.ShareUsers
		if getFile.Hash != file.Hash {
			getFile.HashHistory = append(getFile.HashHistory, &models.HashHistory{
				Hash: getFile.Hash,
			})
			getFile.Hash = file.Hash
		}
		_, err := fd.UpdateFile(getFile)
		if err != nil {
			return nil, err
		}
		return getFile, nil
	}
	if err := file.Default(); err != nil {
		return nil, err
	}

	_, err = file.GetCollection().InsertOne(context.TODO(), file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (fd *FileDbx) SaveStaticFile(sf *models.StaticFile) (*models.StaticFile, error) {
	// 先检测状态正常的有没有
	nfile.IsExists(sf.Path + "/" + sf.FileName)
	if err := sf.Default(); err != nil {
		return nil, err
	}
	// 获取文件实际信息
	if sf.FileInfo.Type == "image/png" || sf.FileInfo.Type == "image/jpeg" {
		imageInfo, err := nimages.GetImageInfo(path.Join(sf.Path, sf.FileName))
		log.Info(imageInfo, err)
		if err != nil {
			return nil, err
		}
		sf.FileInfo.Width = imageInfo.Width
		sf.FileInfo.Height = imageInfo.Height
	}

	_, err := sf.GetCollection().InsertOne(context.TODO(), sf)
	if err != nil {
		return nil, err
	}
	return sf, nil
}

func (fd *FileDbx) GetUnusedStaticFileList(pageSize, pageNum int, deadline int64) ([]*(map[string]interface{}), error) {
	sf := new(models.StaticFile)
	// log.Info("deadline", time.Now().Unix(), deadline)
	// id, _ := primitive.ObjectIDFromHex("64a7cd0c0a4bb98c933df551")
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						// "_id": id,
						"createTime": bson.M{
							"$lte": deadline,
						},
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "File",
				"localField":   "fileInfo.hash",
				"foreignField": "hash",
				"as":           "files",
			},
		},
		{
			"$project": bson.M{
				"_id":      1,
				"path":     1,
				"fileName": 1,
				"files": bson.M{
					"$filter": bson.M{
						"input": "$files",
						"as":    "file",
						"cond": bson.M{
							"$and": bson.M{
								"$gte": bson.A{
									"$$file.status", 0,
								},
							},
						},
					},
				},
			},
		},
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"files": bson.M{
							"$size": 0,
						},
					},
				},
			},
		},
		{
			"$skip": pageSize * (pageNum - 1),
		},
		{
			"$limit": pageSize,
		},
	}

	var results []*(map[string]interface{})
	opts, err := sf.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return results, err
	}
	return results, nil
}

func (fd *FileDbx) DeleteStaticFile(id primitive.ObjectID) error {
	sf := new(models.StaticFile)
	params := bson.M{
		"_id": id,
	}

	_, err := sf.GetCollection().DeleteOne(context.TODO(), params)
	if err != nil {
		return err
	}
	return nil
}
