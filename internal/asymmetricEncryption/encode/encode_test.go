package encode_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/dip96/metrics/internal/asymmetricEncryption/encode"
	"os"
	"testing"

	"github.com/dip96/metrics/internal/config"
)

func TestEncryptData(t *testing.T) {
	// Генерируем новую пару ключей RSA для тестирования
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	publicKey := &privateKey.PublicKey

	// Создаем некоторые тестовые данные для шифрования
	plaintext := []byte("Это тестовые данные для шифрования")

	// Создаем временный файл с открытым ключом
	tmpKeyFile, err := os.CreateTemp("", "test_key.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary key file: %v", err)
	}
	defer os.Remove(tmpKeyFile.Name())

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(publicKey),
	})

	if _, err := tmpKeyFile.Write(pemData); err != nil {
		t.Fatalf("Failed to write public key to temporary file: %v", err)
	}

	// Устанавливаем временный конфиг с путем к файлу с открытым ключом
	cnf := config.LoadAgent()
	cnf.CryptoKey = tmpKeyFile.Name()

	// Вызываем функцию для тестирования
	ciphertext, err := encode.EncryptData(plaintext)
	if err != nil {
		t.Errorf("Failed to encrypt data: %v", err)
	}

	// Расшифровываем зашифрованные данные с помощью закрытого ключа
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		t.Errorf("Failed to decrypt data: %v", err)
	}

	// Проверяем, что расшифрованные данные совпадают с исходными
	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted data does not match original plaintext")
	}
}
