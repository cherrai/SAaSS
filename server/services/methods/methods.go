package methods

import (
	"strings"
	"time"

	"github.com/cherrai/nyanyago-utils/cipher"
)

func GetStaticFilePathAndFileName(hash string, suffix string) (path, fileName string) {
	path = strings.ToLower("./static/storage" +
		"/" + time.Now().Format("2006/01/02"))
	fileName = strings.ToLower(cipher.MD5(hash+time.Now().GoString()) + suffix)
	return
}
