package middleware

import (
	"echo-base-logger/config"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ModifiedContext struct {
	echo.Context
}

func (mc ModifiedContext) Path() string {
	return mc.Request().URL.String()
}

func BaseLogger(conf *config.LoggerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var startTime = time.Now()

			middlewareFunc := middleware.Logger()
			h := middlewareFunc(next)
			n := ModifiedContext{c}

			conf.SetContext(n)

			c.Response().After(func() {
				if c.Response().Status != http.StatusOK || strings.Contains(c.Request().Host, "localhost") {
					buf, _ := conf.DefaultLogger(startTime)
					_, _ = c.Logger().Output().Write(buf)
				}
			})

			if err := h(n); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}
