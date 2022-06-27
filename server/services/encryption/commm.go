package encryption

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"
)

func GetRandomNum(digits int) int64 {
	// fmt.Println("X ", time.Now().UnixNano())
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var digit string = "10"
	for i := 1; i < digits; i++ {
		digit += "0"
	}
	digitNum, err := strconv.ParseInt(digit, 10, 64)
	if err != nil {
		panic(err)
	}
	// fmt.Println("X ", digitNum)
	targetStr := strconv.FormatInt(rnd.Int63n(digitNum), 10)

	for i := len(targetStr); i < digits; i++ {
		randomNum := rand.Intn(10)
		targetStr += strconv.Itoa(randomNum)
	}
	targetNum, err := strconv.ParseInt(targetStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return targetNum
}

func GetRandomKey() string {
	h := md5.New()
	h.Write([]byte(strconv.FormatInt(GetRandomNum(16), 10)))
	return hex.EncodeToString(h.Sum(nil))
}
