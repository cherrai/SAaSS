package methods

import (
	conf "github.com/cherrai/SAaSS/config"
	dbxV1 "github.com/cherrai/SAaSS/dbx/v1"
)

var fileDbx = new(dbxV1.FileDbx)

func DeleteFile() {
	log.Info("------DeleteFile------")
	log.Info(conf.FileExpirationRemovalDeadline)
	log.Info(conf.TempFileRemovalDeadline)

	// 删除临时文件
	// fileDbx.GetAllTempFile
}
