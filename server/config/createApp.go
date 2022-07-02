package conf

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/cipher"
	"github.com/cherrai/nyanyago-utils/nlog"
	"github.com/google/uuid"
)

var (
	log = nlog.New()
)

var (
	AppList   = map[string]typings.AppListItem{}
	ServerId  = ""
	configDir = "./conf.json"
)

func Init() {
	readFile()
	CreateServerId()
	CreateApp()
	writeFile()
}

func readFile() {
	jsonFile, _ := os.Open(configDir)
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)

	conf := new(typings.ServerConfig)
	err := decoder.Decode(conf)
	if err != nil {
		log.Error("Error:", err)
		return
	}
	ServerId = conf.ServerId
	for k, v := range conf.AppList {
		conf.AppList[k] = typings.AppListItem{
			Name:         v.Name,
			AppId:        v.AppId,
			AppKey:       v.AppKey,
			EncryptionId: strings.ToLower(cipher.MD5(v.AppId + "saass")),
		}
	}
	AppList = conf.AppList
	// log.Info("AppList", AppList)
}

func CreateServerId() {
	if ServerId == "" {
		uuid := uuid.New()
		ServerId = uuid.String()
	}
}

func CreateApp() {
	for _, v := range Config.AppList {
		isExists := false
		for _, sv := range AppList {
			if sv.Name == v.Name {
				isExists = true
				break
			}
		}
		if !isExists {
			appId := uuid.New().String()
			appKey := uuid.New().String()
			AppList[appId] = typings.AppListItem{
				Name:         v.Name,
				AppId:        appId,
				EncryptionId: strings.ToLower(cipher.MD5(appId + "saass")),
				AppKey:       appKey,
			}
		}
	}
}

func writeFile() {
	serverConfig, serverConfigErr := os.Create(configDir)
	if serverConfigErr != nil {
		panic(serverConfigErr)
	} else {
		str, _ := json.MarshalIndent(map[string](interface{}){
			"serverId": ServerId,
			"appList":  AppList,
		}, "", "  ")

		_, serverConfigWriteErr := serverConfig.Write([]byte(str))
		if serverConfigWriteErr != nil {
			panic(serverConfigWriteErr)
		}
	}
	defer serverConfig.Close()
}
