package methods

import (
	"io/ioutil"
	"os"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/ntimer"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Clear() {
	ntimer.SetTimeout(func() {

		clear()
		ntimer.SetRepeatTimeTimer(func() {
			clear()
		}, ntimer.RepeatTime{
			Hour: 4,
		}, "Day")
	}, 400)
}

func clear() {
	log.Info("------Clear------")
	clearUnstoredStaticFile("./static/storage")
	clearUnuserdStaticFile(1)
	clearEmptyFolder("./static")
}

// 删除空文件夹
func clearEmptyFolder(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error(err)
	}
	if len(files) == 0 {
		os.RemoveAll(path)
	}
	for _, f := range files {
		if f.IsDir() {
			clearEmptyFolder(path + "/" + f.Name())
		}
	}
}

// 删除没有存储到数据库的静态文件
func clearUnstoredStaticFile(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error(err)
	}
	if len(files) == 0 {
		os.RemoveAll(path)
		return
	}
	for _, f := range files {
		if f.IsDir() {
			clearUnstoredStaticFile(path + "/" + f.Name())
		} else {
			sf, err := fileDbx.GetStaticFileWithPath(path, f.Name())
			if err != nil {
				log.Error(err)
				continue
			}
			if sf == nil {
				log.Info("Remove static file -> ", path+"/"+f.Name())
				os.Remove(path + "/" + f.Name())
			}

		}
	}
}

// 删除未使用的静态文件
func clearUnuserdStaticFile(pageNum int) {
	// log.Info("------clear------")

	// log.Info(conf.Config.File.UnusedFileRetentionTime)
	// log.Info(conf.Config.File.FileTokenSign)
	// 1. 获取所有未用过的文件, 根据创建文件时间筛选遍历超出保留时长的文件

	pageSize := 100

	list, err := fileDbx.GetUnusedStaticFileList(pageSize, pageNum, time.Now().Unix()-conf.Config.File.UnusedFileRetentionTime)
	if err != nil {
		log.Error(err)
	}
	log.Info("len(list)", len(list))

	// 2. 删除这些静态文件的数据库内容和文件内容
	for _, item := range list {
		v := *item
		// log.Info(v)
		files := (v)["files"].(primitive.A)
		// log.Info(files)
		if len(files) > 0 {
			// log.Info("就是你还有")
			// log.Info((v)["_id"], (v)["files"], v["path"], v["fileName"], " -> deleting")
		} else {
			log.Info("该删除了", pageNum)
			// 2.1. 检测静态文件是否存在,有则删除
			path := v["path"].(string) + "/" + v["fileName"].(string)
			if nfile.IsExists(path) {
				os.Remove(path)
				// 2.2 删除数据库内容
				err := fileDbx.DeleteStaticFile(v["_id"].(primitive.ObjectID))
				if err != nil {
					log.Error(err)
				}
			} else {
				log.Info(v["_id"], " -> static file does not exist")
			}
		}
	}
	if len(list) != 0 {
		clearUnuserdStaticFile(pageNum + 1)
	} else {
		log.Info("------end------")
		clearEmptyFolder("./static")
	}
}
