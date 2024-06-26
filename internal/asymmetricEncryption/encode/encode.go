package encode

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/utils"
)

func EncryptData(data []byte) ([]byte, error) {
	cfg, err := config.LoadAgent()
	if err != nil {
		return nil, err
	}

	var keyProvider utils.KeyProvider = utils.RSAKeyProvider{}

	pubKey, err := keyProvider.GetPublicKey(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}

	return rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
}
