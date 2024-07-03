package middleware

import (
	"github.com/dip96/metrics/internal/config"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

func CheckIp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		cfg, err := config.LoadServer()

		if err != nil {
			return err
		}

		// Если trusted_subnet пуст, пропускаем проверку
		if cfg.TrustedSubnet == "" {
			err = next(c)

			return err
		}

		// Получаем IP-адрес из заголовка X-Real-IP
		ipStr := c.Request().Header.Get("X-Real-IP")
		if ipStr == "" {
			log.Error("Not found X-Real-IP")
			return c.HTML(http.StatusForbidden, "")
		}

		// Парсим IP-адрес
		ip := net.ParseIP(ipStr)
		if ip == nil {
			log.Error("Invalid IP address")
			return c.HTML(http.StatusForbidden, "")
		}

		// Парсим доверенную подсеть
		_, trustedIPNet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			log.Error("Invalid trusted subnet configuration")
			return c.HTML(http.StatusForbidden, "")
		}

		// Проверяем, входит ли IP-адрес в доверенную подсеть
		if !trustedIPNet.Contains(ip) {
			log.Error("Untrusted network")
			return c.HTML(http.StatusForbidden, "")
		}

		err = next(c)
		return err
	}
}
