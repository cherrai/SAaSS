package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/nlog"
)

var (
	Log    = nlog.New()
	log    = Log
	Config *typings.Config
	// 文件到期后根据时间进行删除 未做
	// []string{"Image", "Video", "Audio", "Text", "File"}
	FileExpirationRemovalDeadline = 60 * 3600 * 24 * time.Second
	// 临时文件删除期限
	TempFileRemovalDeadline = 60 * 3600 * 24 * time.Second
)

func GetConfig(configPath string) {
	jsonFile, _ := os.Open(configPath)

	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)

	conf := new(typings.Config)
	//Decode从输入流读取下一个json编码值并保存在v指向的值里
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// if Config.Server.Mode == "debug" {
	// 	// Log = nlog.Nil()
	// }
	Config = conf
}
