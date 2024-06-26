package generate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"os"
)

func Generate() {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Получаем публичный ключ из приватного
	publicKey := &privateKey.PublicKey

	// Экспортируем приватный ключ в PEM-формате
	privatePEM, err := pemEncodePrivateKey(privateKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Экспортируем публичный ключ в PEM-формате
	publicPEM, err := pemEncodePublicKey(publicKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	cnf, err := config.LoadServer()

	if err != nil {
		fmt.Printf("Failed to prepare server config: %v\n", err)
		return
	}

	// Записываем ключи в файлы
	err = os.WriteFile(cnf.CryptoKey, privatePEM, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := config.LoadAgent()

	if err != nil {
		fmt.Printf("Failed to prepare agent config: %v\n", err)
		return
	}

	err = os.WriteFile(cfg.CryptoKey, publicPEM, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Ключи успешно созданы и сохранены в файлы private.pem и public.pem")
}

// Кодирует приватный ключ в PEM-формат
func pemEncodePrivateKey(key *rsa.PrivateKey) ([]byte, error) {
	der := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}

// Кодирует публичный ключ в PEM-формат
func pemEncodePublicKey(key *rsa.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(block), nil
}
