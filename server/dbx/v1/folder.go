package dbxV1

import (
	"context"
	"errors"
	"path"
	"strings"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/nyanyago-utils/nshortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FolderDbx struct {
}

func (fd *FolderDbx) NewFolder(appId, folderName, parentPath, authorId string) (*models.Folder, error) {
	// 先检测parentPath是否存在
	log.Info(appId, folderName, parentPath)
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, true, authorId)
	if err != nil {
		return nil, err
	}
	log.Info("parentFolderId", parentFolderId, err)
	// 先检测状态正常的有没有
	gf1, err := fd.GetFolderByParentFolderId(appId, folderName, parentFolderId)
	if err != nil {
		return nil, err
	}
	if gf1 != nil {
		gf1.Status = 1
		fd.UpdateFolderStatus(appId, parentPath, []string{gf1.FolderName}, 1, authorId)
		return gf1, nil
	}
	shortId, err := fd.GetShortId(9)
	if err != nil {
		return nil, err
	}
	// log.Info(parentFolderId, err)
	folder := models.Folder{
		AppId:          appId,
		ShortId:        shortId,
		FolderName:     folderName,
		ParentFolderId: parentFolderId,
		AvailableRange: models.FolderAvailableRange{
			AuthorId: authorId,
		},
	}
	if err := folder.Default(); err != nil {
		return nil, err
	}

	_, err = folder.GetCollection().InsertOne(context.TODO(), folder)
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (u *FolderDbx) GetShortId(digits int) (string, error) {
	str := nshortid.GetSpecifiedRandomString("ABCDEFG", 1) + nshortid.GetShortId(digits)

	// 检测
	file := new(models.Folder)
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

	var results []*models.Folder
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

func (fd *FolderDbx) GetFolderWithShortId(shortId string) (*models.Folder, error) {
	file := new(models.Folder)
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

	var results []*models.Folder
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil {
		// log.Info(results, encryptionName, params)
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("content does not exist")
	}
	return results[0], nil
}

func (fd *FolderDbx) GetParentFolderId(appId, parentPath string, isCreateParentFolder bool, authorId string) (primitive.ObjectID, error) {
	// 先检测parentPath是否存在
	// log.Info(appId, parentPath)

	if parentPath == "/" {
		return primitive.NilObjectID, nil
	}

	parentFolderId := primitive.NilObjectID
	folderName := ""
	parentFolder := ""

	rKey := conf.Redisdb.GetKey("ParentFolderId")

	v, err := conf.Redisdb.Get(rKey.GetKey(parentPath))
	if v != nil && err == nil {
		// log.Info("v.String() redis", parentPath, v.String(), err)
		if v.String() != "" {
			id, _ := primitive.ObjectIDFromHex(v.String())
			return id, nil
		}
	}

	li := strings.LastIndex(parentPath, "/")
	folderName = parentPath[li+1 : len(parentPath)-0]
	parentFolder = path.Join("/", parentPath[:li])
	// log.Info(parentFolder)
	if parentFolder != "/" {
		pfId, err := fd.GetParentFolderId(appId, parentFolder, isCreateParentFolder, authorId)
		if err != nil {
			return primitive.NilObjectID, err
		}
		parentFolderId = pfId
	}

	gf, err := fd.GetFolderByParentFolderId(appId, folderName, parentFolderId)

	// log.Info(gf, err, appId, folderName, parentFolderId)
	if err != nil {
		return primitive.NilObjectID, err
	}

	if gf == nil {
		// 说明此文件夹不存在
		if isCreateParentFolder {
			f, err := fd.NewFolder(appId, folderName, parentFolder, authorId)
			if err != nil {
				return primitive.NilObjectID, err
			}
			return f.Id, nil
		}
		return primitive.NilObjectID, err
	}
	if err = conf.Redisdb.Set(rKey.GetKey(parentPath), gf.Id.Hex(), rKey.GetExpiration()); err != nil {
		return primitive.NilObjectID, err
	}
	return gf.Id, nil
}

func (fd *FolderDbx) GetFolderByParentFolderId(appId, folderName string, parentFolderId primitive.ObjectID) (*models.Folder, error) {
	folder := new(models.Folder)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId":          appId,
						"folderName":     folderName,
						"parentFolderId": parentFolderId,
					},
					// {
					// 	"status": bson.M{
					// 		"$in": []int64{1, 0, -1},
					// 	},
					// },
				},
			},
		},
	}

	var results []*models.Folder
	opts, err := folder.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (fd *FolderDbx) GetFolder(appId, folderName, parentPath string, authorId string) (*models.Folder, error) {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return nil, err
	}
	folder := new(models.Folder)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId":          appId,
						"folderName":     folderName,
						"parentFolderId": parentFolderId,
					},
					// {
					// 	"status": bson.M{
					// 		"$in": []int64{1, 0, -1},
					// 	},
					// },
				},
			},
		},
	}

	var results []*models.Folder
	opts, err := folder.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (fd *FolderDbx) GetFolderList(appId, parentPath string, authorId string) ([]*models.Folder, error) {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return nil, err
	}
	return fd.GetFolderListByParentFolderId(appId, parentFolderId, []int64{1, 0})
}

func (fd *FolderDbx) GetFolderListByParentFolderId(appId string, parentFolderId primitive.ObjectID, status []int64) ([]*models.Folder, error) {
	folder := new(models.Folder)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId":          appId,
						"parentFolderId": parentFolderId,
					},
					{
						"status": bson.M{
							"$in": status,
						},
					},
				},
			},
		},
	}

	var results []*models.Folder
	opts, err := folder.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results, nil
}

func (fd *FolderDbx) SetFileSharing(appId,
	authorId, path string,
	folderNames []string,
	status int64) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.Folder)

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
					"folderName": bson.M{
						"$in": folderNames,
					},
				},
				{
					"status": 1,
				},
			},
		}, bson.M{
			"$set": bson.M{
				"lastUpdateTime":            time.Now().Unix(),
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

func (fd *FolderDbx) SetFolderPassword(appId,
	authorId, path string,
	folderName string,
	password string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, path, false, authorId)
	if err != nil {
		return err
	}
	folder := new(models.Folder)

	updateResult, err := folder.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId": appId,
				},
				{
					"parentFolderId": parentFolderId,
				},
				{
					"folderName": folderName,
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

func (fd *FolderDbx) GetFolderListByAuthorId(
	appId string,
	authorId string,
	pageNum,
	pageSize int64,
	status []int64) ([]*models.Folder, error) {
	folder := new(models.Folder)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId":                   appId,
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

	var results []*models.Folder
	opts, err := folder.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results, nil
}
func (fd *FolderDbx) GetFolderTreeByParentFolderId(appId string, parentFolderId primitive.ObjectID, status []int64) ([]*models.Folder, error) {

	results := []*models.Folder{}

	folders, err := fd.GetFolderListByParentFolderId(appId, parentFolderId, status)
	if err != nil {
		return results, err
	}
	results = append(results, folders...)
	for _, v := range folders {
		sFolders, err := fd.GetFolderTreeByParentFolderId(appId, v.Id, status)
		if err != nil {
			return results, err
		}
		results = append(results, sFolders...)
	}

	return results, nil
}
func (fd *FolderDbx) GetFolderListByFolderNames(appId, parentPath string, folderNames []string, authorId string) ([]*models.Folder, error) {
	folder := new(models.Folder)

	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return nil, err
	}
	log.Info("parentFolderId", parentFolderId)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId":          appId,
						"parentFolderId": parentFolderId,
						"folderName": bson.M{
							"$in": folderNames,
						},
					},
					{
						"status": bson.M{
							"$in": []int64{1, 0},
						},
					},
				},
			},
		},
	}

	var results []*models.Folder
	opts, err := folder.GetCollection().Aggregate(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results, nil
}

func (fd *FolderDbx) UpdateFolderName(appId, parentPath, oldFolderName, newFolderName string, authorId string) error {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return err
	}
	folder := new(models.Folder)
	updateResult, err := folder.GetCollection().UpdateOne(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId":          appId,
					"parentFolderId": parentFolderId,
					"folderName":     oldFolderName,
				},
				{
					"status": bson.M{
						"$in": []int64{1, 0, -1},
					},
				},
			},
		}, bson.M{
			"$set": bson.M{
				"folderName":     newFolderName,
				"lastUpdateTime": time.Now().Unix(),
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

func (fd *FolderDbx) UpdateFolderStatus(appId, parentPath string, folderNames []string, status int64, authorId string) error {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.Folder)
	deleteTime := int64(-1)
	if status == -1 {
		deleteTime = time.Now().Unix()
	}
	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId":          appId,
					"parentFolderId": parentFolderId,
					"folderName": bson.M{
						"$in": folderNames,
					},
				},
			},
		}, bson.M{
			"$set": bson.M{
				"status":         status,
				"lastUpdateTime": time.Now().Unix(),
				"deleteTime":     deleteTime,
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

func (fd *FolderDbx) MoveFoldersToTrash(appId, parentPath string, folderNames []string, authorId string) error {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.Folder)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId":          appId,
					"parentFolderId": parentFolderId,
					"folderName": bson.M{
						"$in": folderNames,
					},
				},
				{
					"status": bson.M{
						"$in": []int64{1, 0},
					},
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
		return errors.New("restore fail")
	}
	return nil
}

func (fd *FolderDbx) Restore(appId, parentPath string, folderNames []string, authorId string) error {
	parentFolderId, err := fd.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return err
	}
	file := new(models.Folder)

	updateResult, err := file.GetCollection().UpdateMany(context.TODO(),
		bson.M{
			"$and": []bson.M{
				{
					"appId":          appId,
					"parentFolderId": parentFolderId,
					"folderName": bson.M{
						"$in": folderNames,
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

func (fd *FolderDbx) CopyFolder(appId, authorId, parentPath string, folderNames []string, newParentPath string) error {
	newParentFolderId, err := folderDbx.GetParentFolderId(appId, newParentPath, false, authorId)
	if err != nil {
		return err
	}

	// 1、先检测目标文件夹是否有现存的
	// log.Info(appId, authorId, parentPath, folderNames, newParentPath)

	folders, err := fd.GetFolderListByFolderNames(appId, parentPath, folderNames, authorId)
	if err != nil {
		return err
	}
	log.Info(folders)

	for _, v := range folders {
		tId := v.Id
		v.Id = primitive.NilObjectID
		v.ParentFolderId = newParentFolderId
		_, err := fd.NewFolder(
			appId,
			v.FolderName,
			newParentPath,
			authorId,
		)
		log.Info(err)
		if err != nil {
			return err
		}

		folders, err := fd.GetFolderList(appId,
			path.Join(parentPath, v.FolderName),
			authorId)
		if err != nil {
			return err
		}
		// log.Info(folder)
		fns := []string{}
		for _, v := range folders {
			fns = append(fns, v.FolderName)
		}

		files, err := fileDbx.GetFileLisByParentFolderId(
			appId,
			tId,
		)
		if err != nil {
			return err
		}
		fileNames := []string{}
		for _, v := range files.List {
			fileNames = append(fileNames, v.FileName)
		}
		log.Info("files", path.Join(parentPath, v.FolderName), files, tId, fileNames)

		err = fileDbx.CopyFile(appId,
			authorId,
			path.Join(parentPath, v.FolderName), fileNames,
			path.Join(newParentPath, v.FolderName),
		)
		if err != nil {
			return err
		}

		err = fd.CopyFolder(appId,
			authorId,
			path.Join(parentPath, v.FolderName), fns,
			path.Join(newParentPath, v.FolderName),
		)
		if err != nil {
			return err
		}

	}
	return nil
}

func (fd *FolderDbx) MoveFolder(appId, authorId, parentPath string, folderNames []string, newParentPath string) error {
	err := fd.CopyFolder(appId, authorId, parentPath, folderNames, newParentPath)
	log.Info("err", err)
	if err != nil {
		return err
	}
	return fd.DeleteFolders(appId, authorId, parentPath, folderNames)
}

func (fd *FolderDbx) DeleteFolders(appId, authorId, parentPath string, folderNames []string) error {
	parentFolderId, err := folderDbx.GetParentFolderId(appId, parentPath, false, authorId)
	if err != nil {
		return err
	}
	folder := new(models.Folder)

	log.Info(parentFolderId, folderNames)

	params := bson.M{
		"$and": []bson.M{
			{
				"appId": appId,
			},
			{
				"parentFolderId": parentFolderId,
			},
			{
				"folderName": bson.M{
					"$in": folderNames,
				},
			},
		},
	}

	for _, v := range folderNames {
		parentFolderId, err := folderDbx.GetParentFolderId(appId, path.Join(parentPath, v), false, authorId)
		if err != nil {
			return err
		}
		folders, err := fd.GetFolderListByParentFolderId(appId, parentFolderId, []int64{1, 0, -1, -2})
		if err != nil {
			return err
		}
		fns := []string{}
		for _, v := range folders {
			fns = append(fns, v.FolderName)
		}

		err = fd.DeleteFolders(appId, authorId, path.Join(parentPath, v), fns)
		if err != nil {
			return err
		}
		files, err := fileDbx.GetFileLisByParentFolderId(appId, parentFolderId)
		if err != nil {
			return err
		}
		fileNames := []string{}
		for _, v := range files.List {
			fileNames = append(fileNames, v.FileName)
		}
		// log.Info(path.Join(parentPath, v), files, fileNames)
		err = fileDbx.DeleteFiles(appId, path.Join(parentPath, v), fileNames, authorId, []int64{1, 0})
		if err != nil {
			return err
		}
	}
	_, err = folder.GetCollection().DeleteMany(context.TODO(), params)
	// log.Info(err)
	if err != nil {
		return err
	}
	rKey := conf.Redisdb.GetKey("ParentFolderId")

	for _, v := range folderNames {
		conf.Redisdb.Delete(rKey.GetKey(path.Join(parentPath, v)))
	}

	return nil
}
