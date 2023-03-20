package conf

import (
	"encoding/json"
	"os"

	"github.com/cherrai/SAaSS/services/typings"
	"github.com/cherrai/nyanyago-utils/nfile"
	"github.com/cherrai/nyanyago-utils/nshortid"
	"github.com/cherrai/nyanyago-utils/nstrings"
	"github.com/google/uuid"
)

var (
	AppList   = map[string]typings.AppListItem{}
	ServerId  = ""
	configDir = "./conf.json"
)

func Init() {
	configDir = Config.AppListDir
	readFile()
	CreateServerId()
	CreateApp()
	writeFile()
}

func readFile() {
	// log.Info("configDir", configDir, nfile.IsExists(configDir))
	if !nfile.IsExists(configDir) {
		writeFile()
	}
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
			EncryptionId: nstrings.StringOr(v.EncryptionId, getShortId(conf.AppList, 9)),
		}
	}
	AppList = conf.AppList
	// log.Info("AppList", AppList)
}

func getShortId(appList map[string]typings.AppListItem, digits int) string {
	id := nshortid.GetShortId(digits)
	flag := false
	for _, v := range appList {
		if v.EncryptionId == id {
			flag = true
			break
		}
	}
	if flag {
		return getShortId(appList, digits)
	}
	return id
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
				EncryptionId: getShortId(AppList, 9),
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
