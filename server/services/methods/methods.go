package methods

import (
	"strings"
	"time"

	"github.com/cherrai/nyanyago-utils/cipher"
	"github.com/cherrai/nyanyago-utils/ncredentials"
)

func GetStaticFilePathAndFileName(hash string, suffix string) (path, fileName string) {
	path = strings.ToLower("./static/storage" +
		"/" + time.Now().Format("2006/01/02"))
	fileName = strings.ToLower(cipher.MD5(hash+time.Now().GoString()) + suffix)
	return
}

func GetTemporaryAccessToken(id string, deadline int64) map[string]string {
	t := time.Duration(deadline-time.Now().Unix()) * time.Second
	user, temporaryAccessToken := ncredentials.GenerateCredentials(id, t)
	return map[string]string{
		"user":                 user,
		"temporaryAccessToken": temporaryAccessToken,
	}
}

func VerfiyTemporaryAccessToken(id, user, temporaryAccessToken string) bool {
	return ncredentials.AuthCredentials(
		user,
		temporaryAccessToken,
		id)
}
