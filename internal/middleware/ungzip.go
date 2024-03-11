package middleware

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func UnzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ce := c.Request().Header.Get("Content-Encoding")
		if ce == "gzip" {
			reader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				log.Error(err)
				return err
			}
			defer reader.Close()
			c.Request().Body = reader
		}
		err := next(c)

		return err
	}
}
