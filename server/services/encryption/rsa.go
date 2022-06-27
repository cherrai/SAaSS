package encryption

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
)

type RsaKey struct {
	PrivateKey []byte
	PublicKey  []byte
}

//RSA公钥私钥产生
func (r *RsaKey) GenerateRsaKey() {
	// 生成私钥文件
	randomNum := rand.Reader
	privateKey, err := rsa.GenerateKey(randomNum, 1024)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	r.PrivateKey = pem.EncodeToMemory(block)

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	r.PublicKey = pem.EncodeToMemory(block)
}

//签名
func (r *RsaKey) RsaSignWithSha256(data []byte) []byte {
	if r.PrivateKey == nil {
		panic(errors.New("The private key does not exist》"))
	}
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(r.PrivateKey)
	if block == nil {
		panic(errors.New("Private key error."))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

func (r *RsaKey) GetRsaSignStringWithSha256(data []byte) (signDataStr string) {
	signData := r.RsaSignWithSha256(data)
	signDataStr = r.SignDataToString(signData)
	return
}

func (r *RsaKey) SignDataToString(signData []byte) string {
	return hex.EncodeToString(signData)
}

//验证
func (r *RsaKey) VerifySignWithSha256(data, signData, publicKeyBytes []byte) bool {
	block, _ := pem.Decode(publicKeyBytes)
	// fmt.Println("block: ", block)
	if block == nil {
		panic(errors.New("[RSA VerifySign]public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	// fmt.Println("pubKey", pubKey, err)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)
	// fmt.Println(hashed)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	// fmt.Println(signData)
	// fmt.Println(err)
	if err != nil {
		panic(err)
	}
	return true
}

// 公钥加密
func (r *RsaKey) Encrypt(data, keyBytes []byte) []byte {
	//解密pem格式的公钥
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("Public key error."))
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

// 私钥加密 未开发完毕
func (r *RsaKey) EncryptWithPrivateKey(data, keyBytes []byte) []byte {
	//解密pem格式的公钥
	block, _ := pem.Decode(r.PrivateKey)
	if keyBytes != nil {
		block, _ = pem.Decode(keyBytes)
	}
	if block == nil {
		panic(errors.New("Public key error."))
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	//加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		panic(err)
	}
	return ciphertext
}

// 私钥解密
func (r *RsaKey) Decrypt(ciphertext, keyBytes []byte) []byte {
	//获取私钥
	block, _ := pem.Decode(r.PrivateKey)
	if keyBytes != nil {
		block, _ = pem.Decode(keyBytes)
	}
	if block == nil {
		panic(errors.New("Private key error!"))
	}
	//解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	// 解密
	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
	if err != nil {
		panic(err)
	}
	return data
}
