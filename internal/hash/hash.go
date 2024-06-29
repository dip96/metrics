package hash

import (
	"crypto/sha256"
	"fmt"
	"github.com/dip96/metrics/internal/config"
)

func CalculateHashAgent(b []byte) string {
	cfg, err := config.LoadAgent()

	if err != nil {
		fmt.Printf("Failed to prepare agent config: %v\n", err)
		return ""
	}

	if cfg.Key != "" {
		calculateHash(b, cfg.Key)
	}

	return ""
}

func CalculateHashServer(b []byte) string {
	cfg, err := config.LoadServer()

	if err != nil {
		fmt.Printf("Failed to prepare server config: %v\n", err)
		return ""
	}

	if cfg.Key != "" {
		return calculateHash(b, cfg.Key)
	}

	return ""
}

func calculateHash(b []byte, key string) string {
	// вычисляем хеш SHA256 от тела запроса и ключа
	hash := sha256.Sum256(append(b, []byte(key)...))
	return fmt.Sprintf("%x", hash[:])
}
