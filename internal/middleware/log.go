package middleware

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"time"
)

func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		if err != nil {
			log.Errorf("Ошибка - %s", err.Error())
		}

		duration := time.Since(start)
		log.Printf("Запрос: %s %s, время - %s", c.Request().URL.Path, c.Request().Method, duration)

		statusCode := c.Response().Status
		responseSize := c.Response().Size
		log.Printf("Ответ: статус код - %d. Размер - %d", statusCode, responseSize)

		return err
	}
}
