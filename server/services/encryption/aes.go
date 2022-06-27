package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func GetAesKey(clientDhKey string, serverDhKey string, randomKey string) string {
	h := md5.New()
	h.Write([]byte(randomKey + clientDhKey + serverDhKey + strconv.FormatInt(time.Now().UnixNano(), 10)))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

type AesEncrypt struct {
	Key  string
	Mode string
}

// 暂时仅CFB模式
// IV需要协商统一，统一由秘钥+每次请求的requestTime时间戳组成吧

//加密字符串
func (a *AesEncrypt) Encrypt(strMesg string, iv string) ([]byte, error) {
	keyByte := []byte(a.Key)
	// 需要协商统一
	var ivByte = keyByte[:aes.BlockSize]
	// fmt.Println("ivByte", ivByte, []byte(iv))
	if iv != "" {
		ivByte = []byte(iv)[:aes.BlockSize]
	}
	encrypted := make([]byte, len(strMesg))
	aesBlockEncrypter, err := aes.NewCipher(keyByte)
	// fmt.Println("aesBlockEncrypter", len(keyByte), aesBlockEncrypter)
	if err != nil {
		return nil, err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, ivByte)
	aesEncrypter.XORKeyStream(encrypted, []byte(strMesg))

	return encrypted, nil
}
func (a *AesEncrypt) EncrypToString(strMesg string, iv string) (string, error) {
	// log.Println("密文(hex)：", hex.EncodeToString(arrEncrypt))
	// log.Println("密文(base64)：", arrEncryptString)

	arrEncrypt, err := a.Encrypt(strMesg, iv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(arrEncrypt), nil
	// return base64.StdEncoding.EncodeToString(arrEncrypt)
}

func (a *AesEncrypt) DecryptWithString(strMesg string) (string, error) {
	value, err := hex.DecodeString(strMesg)
	// log.Println("密文(hex)：", value, err)
	// // log.Println("密文(base64)：", arrEncryptString)

	// arrEncrypt, _ := base64.StdEncoding.DecodeString(strMesg)
	// fmt.Println(arrEncrypt)
	data, err := a.Decrypt(value)
	// fmt.Println(value, data)
	if err != nil {
		return "", err
	}
	return data, nil
}

//解密字符串
func (a *AesEncrypt) Decrypt(src []byte) (strDesc string, err error) {
	keyByte := []byte(a.Key)
	// fmt.Println(len(keyByte))
	var iv = keyByte[:aes.BlockSize]
	// fmt.Println(iv)
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher(keyByte)
	if err != nil {
		return "", err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, src)
	return string(decrypted), nil
}
