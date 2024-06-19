package middleware

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
)

func UnzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ce := c.Request().Header.Get("Content-Encoding")
		headerEncoding := strings.Split(ce, ",")

		for i := len(headerEncoding) - 1; i >= 0; i-- {
			if headerEncoding[i] == "gzip" {
				reader, err := gzip.NewReader(c.Request().Body)
				if err != nil {
					log.Error(err)
					return err
				}
				defer func(reader *gzip.Reader) {
					err := reader.Close()
					if err != nil {
						//достоточно ли просто лога о том, что была ошибка в defer?
						//или же нужны дополнительные маципуляции в данном случаи?
						log.Error("Reader.Close - error")
					}
				}(reader)
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						//достоточно ли просто лога о том, что была ошибка в defer?
						//или же нужны дополнительные маципуляции в данном случаи?
						log.Error("Body.Close - error")
					}
				}(c.Request().Body)
				c.Request().Body = reader
			}
		}

		err := next(c)

		return err
	}
}
