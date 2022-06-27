package dbxV1

import (
	"context"
	"time"

	"github.com/cherrai/SAaSS/models"
	"github.com/cherrai/nyanyago-utils/nimages"
	"github.com/cherrai/nyanyago-utils/nlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	log = nlog.New()
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
		}, options.Update().SetUpsert(false))

	if err != nil {
		return err
	}
	return nil
}

func (fd *FileDbx) GetFileWithEncryptionName(encryptionName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"encryptionName": encryptionName,
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
				},
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
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
				},
			},
		},
	}

	var results []*models.File
	opts, err := file.GetCollection().Aggregate(context.TODO(), params)
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (fd *FileDbx) GetFileWithFileInfo(appId string, path string, fileName string) (*models.File, error) {
	file := new(models.File)
	params := []bson.M{
		{
			"$match": bson.M{
				"$and": []bson.M{
					{
						"appId": appId,
					},
					{
						"path": path,
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
	if err = opts.All(context.TODO(), &results); err != nil || len(results) == 0 {
		return nil, err
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
				"staticFolderPath":              file.StaticFolderPath,
				"staticFileName":                file.StaticFileName,
				"status":                        file.Status,
				"deleteTime":                    file.DeleteTime,
				"availableRange.visitCount":     file.AvailableRange.VisitCount,
				"availableRange.expirationTime": file.AvailableRange.ExpirationTime,
			},
		}, options.Update().SetUpsert(false))

	if err != nil {
		return nil, err
	}
	return result, nil
}

func (fd *FileDbx) SaveFile(file *models.File) (*models.File, error) {
	// 先检测状态正常的有没有
	getFile, err := fd.GetFileWithFileInfo(file.AppId, file.Path, file.FileName)
	log.Info(getFile, err)
	if err != nil {
		return nil, err
	}
	if getFile != nil {
		getFile.Status = 1
		getFile.DeleteTime = -1
		getFile.StaticFolderPath = file.StaticFolderPath
		getFile.StaticFileName = file.StaticFileName
		getFile.AvailableRange.VisitCount = file.AvailableRange.VisitCount
		getFile.AvailableRange.ExpirationTime = file.AvailableRange.ExpirationTime
		_, err := fd.UpdateFile(getFile)
		if err != nil {
			return nil, err
		}
		return getFile, nil
	}
	if err := file.Default(); err != nil {
		return nil, err
	}
	// 获取文件实际信息
	switch file.Type {
	case "Image":
		imageInfo, err := nimages.GetImageInfo(file.StaticFolderPath + file.StaticFileName)
		log.Info(imageInfo, err)
		if err != nil {
			return nil, err
		}
		file.FileInfo.Width = imageInfo.Width
		file.FileInfo.Height = imageInfo.Height
		break

	}

	_, err = file.GetCollection().InsertOne(context.TODO(), file)
	if err != nil {
		return nil, err
	}
	return file, nil
}
