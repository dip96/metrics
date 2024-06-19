package middleware

import (
	"bytes"
	"github.com/dip96/metrics/internal/asymmetricEncryption/decode"
	"github.com/labstack/echo/v4"
	"io"
	"strings"
)

func DecodeMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ce := c.Request().Header.Get("Content-Encoding")
		headerEncoding := strings.Split(ce, ",")

		// Создаем буфер для чтения тела запроса
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		// Новый входной поток из буфера для передачи в следующий обработчик
		c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		for i := len(headerEncoding) - 1; i >= 0; i-- {
			if headerEncoding[i] == "encrypted" {
				// Расшифровываем данные
				data2, err := decode.DecryptData(body)
				if err != nil {
					return err
				}

				c.Request().Body = io.NopCloser(bytes.NewBuffer(data2))
			}
		}

		err = next(c)

		return err
	}
}
