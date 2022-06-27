package typings

type ServerConfig struct {
	ServerId string
	AppList  map[string]AppListItem
}

type Config struct {
	Server           Server
	AppList          []AppListItem
	StaticPathDomain string
	SSO              Sso
	Redis            Redis
	Mongodb          Mongodb
	StaticUrlPrefix  string
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
type Sso struct {
	AppId  string
	AppKey string
	Host   string
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
	Ssodb struct {
		Name string
		Uri  string
	}
}
