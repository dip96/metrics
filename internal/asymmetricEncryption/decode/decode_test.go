package decode

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/dip96/metrics/internal/config"
)

func TestDecryptData(t *testing.T) {
	// Создаем новый закрытый и открытый ключи RSA для тестирования
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Создаем некоторые тестовые данные для шифрования
	plaintext := []byte("Тестовые данные для шифрования")

	// Зашифровываем данные с помощью открытого ключа
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Устанавливаем временный путь к файлу с закрытым ключом
	tmpKeyFile, err := os.CreateTemp("", "test_key.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary key file: %v", err)
	}
	defer os.Remove(tmpKeyFile.Name())

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	if _, err := tmpKeyFile.Write(pemData); err != nil {
		t.Fatalf("Failed to write private key to temporary file: %v", err)
	}

	// Устанавливаем временный конфиг с путем к файлу с закрытым ключом
	cnf, err := config.LoadServer()

	if err != nil {
		t.Fatalf("Failed to prepare server config: %v\n", err)
	}

	cnf.CryptoKey = tmpKeyFile.Name()

	// Вызываем функцию для тестирования
	decrypted, err := DecryptData(ciphertext)
	if err != nil {
		t.Errorf("Failed to decrypt data: %v", err)
	}

	// Проверяем, что расшифрованные данные совпадают с исходными
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted data does not match original plaintext")
	}
}
