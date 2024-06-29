package middleware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/dip96/metrics/internal/config"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func CheckHash(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg, err := config.LoadServer()

		if err != nil {
			return err
		}

		if cfg.Key != "" {
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				log.Error("Error with read all")
				return c.JSONBlob(http.StatusBadRequest, body)
			}

			c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

			// Вычисление хеша SHA256 от тела запроса и ключа
			expected := sha256.Sum256(append(body, []byte(cfg.Key)...))
			expectedHash := fmt.Sprintf("%x", expected[:])

			receivedHash := c.Request().Header.Get("HashSHA256")

			if receivedHash != expectedHash {
				log.Error("Different hash")
				return c.JSON(http.StatusBadRequest, "Invalid hash")
			}
		}

		err = next(c)

		return err
	}
}
