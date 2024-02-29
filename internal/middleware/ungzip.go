package middleware

import (
	"compress/gzip"
	"github.com/labstack/echo/v4"
)

func UnzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ce := c.Request().Header.Get("Content-Encoding")
		if ce == "gzip" {
			reader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return err
			}
			defer reader.Close()
			c.Request().Body = reader
		}
		err := next(c)

		//if err := next(c); err != nil {
		//	return err
		//}

		//ae := c.Request().Header.Get("Accept-Encoding")

		//if ae == "gzip" {
		//	//_ := c.Response()
		//	//rw := response.Writer()
		//	//body, err := io.ReadAll(rw.Body)
		//	//if err != nil {
		//	//	return err
		//	//}
		//}

		c.Response().Header().Set("Content-Encoding", "gzip")
		c.Response().Header().Set("Accept-Encoding", "gzip")

		return err
	}
}
