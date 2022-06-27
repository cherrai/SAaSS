package encryption

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type EncryptionOption struct {
	RsaKey          *RsaKey
	AesKey          string
	PublicAesKey    *AesEncrypt
	RedisClient     *redis.Client
	RsaKeyDelayDays int64
	UserAesKeyMins  int64
	TempDataMins    int64
}

var ctx = context.Background()

type NewOption struct {
	RedisClient     *redis.Client
	RsaKeyDelayDays int64
	UserAesKeyMins  int64
	TempDataMins    int64
}

func New(opt NewOption) *EncryptionOption {
	if opt.RedisClient == nil {
		panic(errors.New("NewOption RedisClient is empty."))
	}
	// 获取RSA Key
	rsaKey := new(RsaKey)

	// // 获取DH KEY
	// dhKey := new(DhKey)
	// InitDhKey(dhKey)
	eo := EncryptionOption{
		RsaKey:          rsaKey,
		RedisClient:     opt.RedisClient,
		RsaKeyDelayDays: opt.RsaKeyDelayDays,
		UserAesKeyMins:  opt.UserAesKeyMins,
		TempDataMins:    opt.TempDataMins,
	}
	eo.Init()

	return &eo
	// publicFile, err := os.Create("public.pem")
	// defer publicFile.Close()
}

func InitDhKey(dhKey *DhKey) {
	dhKey.CreateDiffieHellman()
}

// 暂定Redis只存储10分钟，如果没有再重新从文件内获取
// 如果文件内10天时间到，就直接重新生成
func (e *EncryptionOption) Init() {
	// 有效期10天
	var delayDays int64 = e.RsaKeyDelayDays
	os.Mkdir("./certs", 0777)

	publicAesKeyContent, err := ioutil.ReadFile("./certs/publicAesKey.key")
	if err != nil {
		fmt.Println(err)
	}
	generateInfoContent, err := ioutil.ReadFile("./certs/generate.info")
	if err != nil {
		fmt.Println(err)
	}

	privateKeyContent, err := ioutil.ReadFile("./certs/private.key")
	if err != nil {
		fmt.Println(err)
	}

	publicKeyContent, err := ioutil.ReadFile("./certs/public.pem")
	if err != nil {
		fmt.Println(err)
	}
	generateInfoInt, err := strconv.ParseInt(string(generateInfoContent), 10, 64)
	// fmt.Println(generateInfoInt < time.Now().Unix() || len(publicAesKeyContent) == 0 || len(privateKeyContent) == 0 || len(publicKeyContent) == 0)
	if generateInfoInt < time.Now().Unix() || len(publicAesKeyContent) == 0 || len(privateKeyContent) == 0 || len(publicKeyContent) == 0 {
		e.RsaKey.GenerateRsaKey()
		rsaPrivateKeyFile, rsaPrivateKeyFileErr := os.Create("./certs/private.key")
		defer rsaPrivateKeyFile.Close()
		if rsaPrivateKeyFileErr != nil {
			panic(rsaPrivateKeyFileErr)
		} else {
			_, rsaPrivateKeyFileWriteErr := rsaPrivateKeyFile.Write([]byte(e.RsaKey.PrivateKey))
			if rsaPrivateKeyFileWriteErr != nil {
				panic(rsaPrivateKeyFileWriteErr)
			}
		}

		rsaPuclicKeyFile, rsaPuclicKeyFileErr := os.Create("./certs/public.pem")
		defer rsaPuclicKeyFile.Close()
		if rsaPuclicKeyFileErr != nil {
			panic(rsaPuclicKeyFileErr)
		} else {
			_, rsaPuclicKeyFileWriteErr := rsaPuclicKeyFile.Write([]byte(e.RsaKey.PublicKey))
			if rsaPuclicKeyFileWriteErr != nil {
				panic(rsaPuclicKeyFileWriteErr)
			}
		}

		generateInfoFile, generateInfoFileErr := os.Create("./certs/generate.info")
		defer generateInfoFile.Close()
		if generateInfoFileErr != nil {
			panic(generateInfoFileErr)
		} else {
			_, generateInfoFileWriteErr := generateInfoFile.Write([]byte(fmt.Sprint(time.Now().Unix() + delayDays*3600*24)))
			if generateInfoFileWriteErr != nil {
				panic(generateInfoFileWriteErr)
			}
		}

		publicAesKeyInfo, publicAesKeyInfoErr := os.Create("./certs/publicAesKey.key")
		defer publicAesKeyInfo.Close()
		if publicAesKeyInfoErr != nil {
			panic(publicAesKeyInfoErr)
		} else {
			randomKey := GetRandomKey()
			fmt.Println("randomKey", randomKey)
			e.PublicAesKey = &AesEncrypt{
				Key:  randomKey,
				Mode: "CFB",
			}
			// 这里生成
			_, publicAesKeyFileWriteErr := publicAesKeyInfo.Write([]byte(randomKey))
			if publicAesKeyFileWriteErr != nil {
				panic(publicAesKeyFileWriteErr)
			}
		}
	} else {
		e.RsaKey.PrivateKey = privateKeyContent
		e.RsaKey.PublicKey = publicKeyContent
		e.PublicAesKey = &AesEncrypt{
			Key:  string(publicAesKeyContent),
			Mode: "CFB",
		}
	}
}
func (e *EncryptionOption) GenerateRandomKey(key string) (randomKey string, err error) {
	// 生成RandomKey
	var delayMins time.Duration = time.Duration(e.TempDataMins)
	randomKey = GetRandomKey()
	err = e.RedisClient.Set(ctx, "encryption_randomkey_"+key, randomKey, delayMins*60*time.Second).Err()
	return
}
func (e *EncryptionOption) GetRandomKey(key string) (randomKey string, err error) {
	// 获取RandomKey
	randomKey, err = e.RedisClient.Get(ctx, "encryption_randomkey_"+key).Result()
	if randomKey == "" {
		randomKey, err = e.GenerateRandomKey(key)
	}
	return
}
func (e *EncryptionOption) GetDhKey(key string) (dhKey DhKey, err error) {
	var delayMins time.Duration = time.Duration(e.TempDataMins)
	baseKey := "encryption_dhkey_"
	var dhKeyStr string
	dhKeyStr, err = e.RedisClient.Get(ctx, baseKey+key).Result()
	if dhKeyStr == "" {
		dhKey.CreateDiffieHellman()
		dhKeyStr, _ := json.Marshal(dhKey)
		err = e.RedisClient.Set(ctx, baseKey+key, dhKeyStr, delayMins*60*time.Second).Err()
	} else {
		err = json.Unmarshal([]byte(dhKeyStr), &dhKey)
		if err != nil {
			return
		}
	}
	return
}

// 存储AES Key以UID为单位到Redis
// 暂定10分钟
func (e *EncryptionOption) GetUserAesKey(key string) (string, error) {
	baseKey := "encryption_user_aeskey_"
	// fmt.Println("e.RedisClient", e.RedisClient)
	// fmt.Println("get", baseKey+key)
	aesKey, err := e.RedisClient.Get(ctx, baseKey+key).Result()
	// fmt.Println("err", err)
	if err != nil {
		return "", err
	}
	return aesKey, nil

}
func (e *EncryptionOption) GetUserAesKeyWithAesKey(key string) (string, error) {
	deUId, deUIDErr := e.PublicAesKey.DecryptWithString(key)
	deUIdSlice := strings.Split(deUId, ".")

	if len(deUIdSlice) != 2 {
		return "", errors.New("Decryption failed.")
	}
	uid := deUIdSlice[1]
	if deUIDErr != nil {
		return "", deUIDErr
	}
	getAesKey, aesKeyErr := e.GetUserAesKey(uid)
	if aesKeyErr != nil {
		return "", aesKeyErr
	}
	return getAesKey, nil
}
func (e *EncryptionOption) SetUserAesKey(sharePublicKeyInt64 int64, key string) (string, error) {
	var delayMins time.Duration = time.Duration(e.UserAesKeyMins)
	baseKey := "encryption_user_aeskey_"

	// 2、获取RandomKey
	randomKey, randomKeyErr := e.GetRandomKey(key)
	if randomKeyErr != nil {
		return "", randomKeyErr
	}
	dhKey, dhKeyErr := e.GetDhKey(key)
	if dhKeyErr != nil {
		return "", dhKeyErr
	}
	aesKey := dhKey.GenerateKey(big.NewInt(sharePublicKeyInt64))
	h := md5.New()
	h.Write([]byte(aesKey + randomKey))
	md5Key := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	// fmt.Println("baseKey+key", baseKey+key)
	err := e.RedisClient.Set(ctx, baseKey+key, md5Key, delayMins*60*time.Second).Err()
	if err != nil {
		return "", err
	}

	return md5Key, nil
}

func Test() {
	// 1、服务端生成RSA公钥和私钥，存储私钥到缓存，
	// 有效期5分钟。将公钥给客户端。
	// fmt.Println(string(serverRsaKey.PrivateKey))

	var serverRsaKey RsaKey
	serverRsaKey.GenerateRsaKey()

	fmt.Println(serverRsaKey)
	fmt.Println(string(serverRsaKey.PublicKey))

	var serverRsaPublicKey = string(serverRsaKey.PublicKey)
	serverRsaPublicKeySignData := serverRsaKey.GetRsaSignStringWithSha256([]byte(serverRsaPublicKey))
	fmt.Println("serverRsaPublicKeySignData ", serverRsaPublicKeySignData)

	// 2、客户端拿到服务端RSA公钥后，生成自己的
	// RSA公钥和私钥，并用服务端RSA公钥加密客户
	// 端公钥传给服务端。有效期5分钟。
	var clientRsaKey RsaKey
	clientRsaKey.GenerateRsaKey()
	// fmt.Println(string(clientRsaKey.PrivateKey))
	// fmt.Println(string(clientRsaKey.PublicKey))

	var clientRsaPublicKey = string(clientRsaKey.PublicKey)
	clientRsaPublicKeySignData := clientRsaKey.GetRsaSignStringWithSha256([]byte(clientRsaPublicKey))
	fmt.Println("serverRsaPublicKeySignData ", clientRsaPublicKeySignData)

	// 3、服务端通过DH算法将生成的PGA，并且再生
	// 成一个用于二次加密的二次Key。用客户端公钥
	// 加密后传给客户端。
	// p为16位质数
	var serverDhKey DhKey
	serverDhKey.CreateDiffieHellman()
	randomKey := GetRandomKey()
	fmt.Println("randomKey: ", randomKey)
	fmt.Println("serverDhKey: ", serverDhKey)
	// 4、客户端用服务端的PGA利用DH算法，生成
	// AES Key和自己的PGA，再通过客户端RSA公钥
	// 将客户端的PGA加密传给服务端。此时的AES Key
	// 混淆一下，并通过之前服务端的二次Key进行加密
	// 后作为真正的AES Key。	}
	var clientDhKey DhKey = DhKey{
		Prime:            serverDhKey.Prime,
		Base:             serverDhKey.Base,
		OutsidePublicKey: serverDhKey.InsidePublicKey,
	}
	clientDhKey.CreateDiffieHellman()
	clientDhKey.GenerateKey(nil)
	serverDhKey.GenerateKey(clientDhKey.InsidePublicKey)
	fmt.Println("serverDhKey: ", serverDhKey)
	fmt.Println("serverDhKey Key: ", serverDhKey.Key)

	fmt.Println("clientDhKey: ", clientDhKey)
	fmt.Println("clientDhKey Key: ", clientDhKey.Key)

	// fmt.Println(
	// 	"Prime:", serverDhKey.Prime,
	// 	"Base:", serverDhKey.Base,
	// 	"A:", serverDhKey.A,
	// 	"B:", serverDhKey.B,
	// 	"randNum:", serverDhKey.randNum,
	// 	"Key:", serverDhKey.Key)
	// fmt.Println(
	// 	"Prime:", clientDhKey.Prime,
	// 	"Base:", clientDhKey.Base,
	// 	"A:", clientDhKey.A,
	// 	"B:", clientDhKey.B,
	// 	"randNum:", clientDhKey.randNum,
	// 	"Key:", clientDhKey.Key)
	// 5、服务端拿到的客户端PGA生成同样的AES Key
	// 之后，Redis缓存5分钟。此时的AES Key混淆
	// 一下，并通过之前服务端的二次Key进行加密后作
	// 为真正的AES Key。互相就用这个加密和解密请
	// 求响应热数据。
	// aesKey := GetAesKey(clientDhKey.Key, serverDhKey.Key, randomKey)
	// fmt.Println("aesKey: ", aesKey, len(aesKey))

	// 6、在未登录的时候。因为没有token，可以直接
	// 走上面的流程。但是登录的时候需要有AES Key加
	// 密的password作为自定义条件获取token。
	// 7、除了登录请求外，其他的请求都必须由token
	// 通过认证且由其包含作为key加密后才能进行请求。
	// 譬如第一次生成客户端公钥私钥，需要包含token
	// 里面解析出来的随机数据。且token必须有效。

	// 	aesEnc := AesEncrypt{
	// 		Key:  aesKey,
	// 		Mode: "CFB",
	// 	}
	// 	arrEncrypt := aesEnc.EncrypToString("{code:200,msg:'Resquest Success',}")
	// 	fmt.Println("arrEncrypt", arrEncrypt)
	// 	strMsg := aesEnc.DecryptWithString(arrEncrypt)
	// 	fmt.Println("strMsg: ", strMsg)
}
