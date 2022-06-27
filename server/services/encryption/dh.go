package encryption

import (
	"math/big"
)

type DhKey struct {
	Prime            *big.Int
	Base             *big.Int
	RandNum          *big.Int
	InsidePublicKey  *big.Int
	OutsidePublicKey *big.Int
	Key              string
}

// 初始化
func (d *DhKey) CreateDiffieHellman() {
	if d.Prime != nil && d.Base != nil {
		d.RandNum = big.NewInt(GetRandomNum(2))
		d.InsidePublicKey = big.NewInt(0).Exp(d.Base, d.RandNum, d.Prime)
	} else {
		d.Prime = big.NewInt(GetRandomNum(16))
		d.Base = big.NewInt(GetRandomNum(16))
		d.RandNum = big.NewInt(GetRandomNum(2))
		d.InsidePublicKey = big.NewInt(0).Exp(d.Base, d.RandNum, d.Prime)
	}
}

// g *big.Int, p *big.Int, A *big.Int
func (d *DhKey) GenerateKey(publicKey *big.Int) string {
	if publicKey == nil {
		// 作为客户端
		SB := big.NewInt(0).Exp(d.OutsidePublicKey, d.RandNum, d.Prime)
		d.Key = SB.String()
	} else {
		// 作为服务端
		d.OutsidePublicKey = publicKey
		SA := big.NewInt(0).Exp(d.OutsidePublicKey, d.RandNum, d.Prime)
		d.Key = SA.String()
	}

	return d.Key
}
