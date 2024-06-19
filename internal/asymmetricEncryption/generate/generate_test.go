package generate

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dip96/metrics/internal/config"
)

func TestGenerate(t *testing.T) {
	// Создаем временные файлы для хранения ключей
	tempDir, err := ioutil.TempDir("", "test-keys")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")

	// Устанавливаем временный конфиг с путями к файлам для ключей
	cnf := config.LoadServer()
	cnf.CryptoKey = privateKeyPath
	// Устанавливаем временный конфиг с путем к файлу с открытым ключом
	cnfAgent := config.LoadAgent()
	cnfAgent.CryptoKey = publicKeyPath

	// Вызываем функцию генерации ключей
	Generate()

	// Проверяем, что файлы с ключами были созданы
	_, err = os.Stat(privateKeyPath)
	if err != nil {
		t.Errorf("Failed to create private key file: %v", err)
	}

	_, err = os.Stat(publicKeyPath)
	if err != nil {
		t.Errorf("Failed to create public key file: %v", err)
	}

	// Проверяем, что содержимое файлов соответствует ожидаемому формату
	privateKeyData, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		t.Errorf("Failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		t.Error("Failed to decode PEM block containing private key")
	}

	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse private key: %v", err)
	}

	publicKeyData, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		t.Errorf("Failed to read public key file: %v", err)
	}

	block, _ = pem.Decode(publicKeyData)
	if block == nil {
		t.Error("Failed to decode PEM block containing public key")
	}

	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Errorf("Failed to parse public key: %v", err)
	}
}
