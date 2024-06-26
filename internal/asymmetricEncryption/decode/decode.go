package decode

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/dip96/metrics/internal/config"
	"github.com/dip96/metrics/internal/utils"
)

func DecryptData(ciphertext []byte) ([]byte, error) {
	cnf, err := config.LoadServer()
	if err != nil {
		return nil, err
	}

	var keyProvider utils.KeyProvider = utils.RSAKeyProvider{}

	privateKey, err := keyProvider.GetPrivateKey(cnf.CryptoKey)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
}
