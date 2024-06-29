package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// KeyProvider интерфейс для получения ключей
type KeyProvider interface {
	GetPrivateKey(path string) (*rsa.PrivateKey, error)
	GetPublicKey(path string) (*rsa.PublicKey, error)
}

// RSAKeyProvider реализует интерфейс KeyProvider для RSA ключей
type RSAKeyProvider struct{}

func (p RSAKeyProvider) GetPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemData, err := readPemFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(pemData)
}

func (p RSAKeyProvider) GetPublicKey(path string) (*rsa.PublicKey, error) {
	pemData, err := readPemFile(path)
	if err != nil {
		return nil, err
	}
	return parsePublicKey(pemData)
}

func readPemFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func parsePrivateKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parsePublicKey(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
