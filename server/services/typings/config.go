package typings

type ServerConfig struct {
	ServerId string
	AppList  map[string]AppListItem
}

type Config struct {
	Server           Server
	AppList          []AppListItem
	FileTokenSign    string
	StaticPathDomain string
	Redis            Redis
	Mongodb          Mongodb
}

type AppListItem struct {
	Name         string `json:"name"`
	AppId        string `json:"appId"`
	AppKey       string `json:"appKey"`
	EncryptionId string `json:"encryptionId"`
}
type Server struct {
	Port int
	Cors struct {
		AllowOrigins []string
	}
	// mode: release debug
	Mode string
}
type Redis struct {
	Addr     string
	Password string
	DB       int
}
type Mongodb struct {
	Currentdb struct {
		Name string
		Uri  string
	}
}
