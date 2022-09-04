package methods

import (
	"io/ioutil"
	"os"
	"time"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/ntimer"
)

func Clear() {
	ntimer.SetTimeout(func() {
		clearEmptyFolder("./static")
		clearUnstoredStaticFile("./static/storage")
		clearUnuserdFile(1)
		ntimer.SetRepeatTimeTimer(func() {
			clearEmptyFolder("./static")
			clearUnstoredStaticFile("./static/storage")
			clearUnuserdFile(1)
		}, ntimer.RepeatTime{
			Hour: 4,
		}, "Day")
	}, 400)
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
				os.Remove(path + "/" + f.Name())
			}

		}
	}
}

// 删除未使用的文件
func clearUnuserdFile(pageNum int) {
	// log.Info("------clear------")

	// log.Info(conf.Config.File.UnusedFileRetentionTime)
	// log.Info(conf.Config.File.FileTokenSign)
	// 1. 获取所有未用过的文件, 根据创建文件时间筛选遍历超出保留时长的文件

	pageSize := 100

	list, err := fileDbx.GetUnusedStaticFileList(pageSize, pageNum, time.Now().Unix()-conf.Config.File.UnusedFileRetentionTime)
	if err != nil {
		log.Error(err)
	}
	log.Info(len(list))

	// 2. 删除这些静态文件的数据库内容和文件内容
	for _, v := range list {
		log.Info(v.Id, " -> deleting")
		// 2.1. 检测静态文件是否存在,有则删除
		path := v.Path + "/" + v.FileName
		if nfile.IsExists(path) {
			os.Remove(path)
		} else {
			log.Info(v.Id, " -> static file does not exist")
		}
		// 2.2 删除数据库内容
		err := fileDbx.DeleteStaticFile(v.Id)
		if err != nil {
			log.Error(err)
		}
	}
	if len(list) != 0 {
		clearUnuserdFile(pageNum + 1)
	} else {
		log.Info("------end------")
	}
}
