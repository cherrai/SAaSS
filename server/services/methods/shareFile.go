package methods

import (
	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/models"
	"path/filepath"
	"time"
)

func FormatFile(list []*models.File) ([]map[string]interface{}, error) {
	listMap := []map[string]interface{}{}
	hashList := []string{}
	for _, v := range list {
		hashList = append(hashList, v.Hash)
	}

	staticFileList, err := fileDbx.GetStaticFileListWithHash(hashList)
	if err != nil {
		return listMap, err
	}
	for _, v := range list {
		for _, sv := range staticFileList {
			if sv.FileInfo.Hash == v.Hash {
				at := GetTemporaryAccessToken(v.ShortId, time.Now().Add(15*60*time.Second).Unix())
				shortUrl := "/s/" + v.ShortId + "?u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
				url := "/s/" + v.FileName + "?sid=" + v.ShortId + "&u=" + at["user"] + "&tat=" + at["temporaryAccessToken"]
				listMap = append(listMap, map[string]interface{}{
					"id":       v.Id,
					"shortId":  v.ShortId,
					"fileName": v.FileName,
					// "path":           params.Path,
					"parentFolderId": v.ParentFolderId.Hex(),
					"availableRange": map[string]interface{}{
						"visitCount":     v.AvailableRange.VisitCount,
						"expirationTime": v.AvailableRange.ExpirationTime,
						"password":       v.AvailableRange.Password,
						"allowShare":     v.AvailableRange.AllowShare,
						"shareUsers":     v.AvailableRange.ShareUsers,
						"authorId":       v.AvailableRange.AuthorId,
					},
					"usage": map[string]interface{}{
						"visitCount": v.Usage.VisitCount,
					},
					"createTime":     v.CreateTime,
					"lastUpdateTime": v.LastUpdateTime,
					"deleteTime":     v.DeleteTime,
					"urls": map[string]string{
						"domainUrl": conf.Config.StaticPathDomain,
						"shortUrl":  shortUrl,
						"url":       url,
					},
					"fileInfo": map[string]interface{}{
						"name":         sv.FileInfo.Name,
						"size":         sv.FileInfo.Size,
						"type":         sv.FileInfo.Type,
						"suffix":       sv.FileInfo.Suffix,
						"lastModified": sv.FileInfo.LastModified,
						"hash":         sv.FileInfo.Hash,
						"width":        sv.FileInfo.Width,
						"height":       sv.FileInfo.Height,
					},
				})
			}
		}
	}
	return listMap, nil
}

func FormatFolder(apiUrl, sid, pwd, parentPath string, list []*models.Folder) ([]map[string]interface{}, error) {
	listMap := []map[string]interface{}{}

	for _, v := range list {
		// at := GetTemporaryAccessToken(v.ShortId, time.Now().Add(15*60*time.Second).Unix())
		url := apiUrl + "?path=" + filepath.Join(parentPath, v.FolderName) + "&sid=" + sid + "&pwd=" + pwd

		listMap = append(listMap, map[string]interface{}{
			"id":             v.Id,
			"folderName":     v.FolderName,
			"shortId":        v.ShortId,
			"parentFolderId": v.ParentFolderId,
			"path":           parentPath,
			"status":         v.Status,
			"availableRange": map[string]interface{}{
				"password":   v.AvailableRange.Password,
				"allowShare": v.AvailableRange.AllowShare,
				"shareUsers": v.AvailableRange.ShareUsers,
				"authorId":   v.AvailableRange.AuthorId,
			},
			"urls": map[string]string{
				"domainUrl": conf.Config.StaticPathDomain,
				"url":       url,
			},
			"usage":          map[string]interface{}{},
			"createTime":     v.CreateTime,
			"lastUpdateTime": v.LastUpdateTime,
			"deleteTime":     v.DeleteTime,
		})
	}
	return listMap, nil
}
