package middleware

import (
	"bytes"
	"github.com/modanisatech/echo-base-logger/config"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

var outputBuffer = &bytes.Buffer{}

func _init() {
	env := os.Getenv("APP_ENV")
	middleware.DefaultLoggerConfig.Format = fmt.Sprintf(`{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",`+
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",`+
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"`+
		`,"environment":"%s","bytes_in":${bytes_in},"bytes_out":${bytes_out}}`, env) + "\n"
	middleware.DefaultLoggerConfig.Skipper = func(context echo.Context) bool {
		return true
	}
}

func TestShouldWriteLogIntoStdOutIfStatusIsNotOK_BaseLogger(t *testing.T) {
	_init()
	loggerConfig := config.NewLoggerConfig(&middleware.DefaultLoggerConfig)
	e := echo.New()
	h := BaseLogger(loggerConfig)(func(context echo.Context) error {
		context.Logger().SetOutput(outputBuffer)
		return nil
	})

	defer outputBuffer.Reset()

	req := httptest.NewRequest(echo.GET, "/", nil)
	res := httptest.NewRecorder()

	c := e.NewContext(req, res)

	if assert.NoError(t, h(c)) {
		if assert.NoError(t, baseLoggerStatusNotFound(c)) {
			assert.Equal(t, http.StatusNotFound, res.Code)
			assert.Contains(t, outputBuffer.String(), `"status":404`)
		}
	}
}

func TestShouldNotWriteLogIntoStdOutIfStatusIsOK_BaseLogger(t *testing.T) {
	_init()
	loggerConfig := config.NewLoggerConfig(&middleware.DefaultLoggerConfig)
	e := echo.New()
	h := BaseLogger(loggerConfig)(func(context echo.Context) error {
		context.Logger().SetOutput(outputBuffer)
		return nil
	})

	defer outputBuffer.Reset()

	req := httptest.NewRequest(echo.GET, "/", nil)
	res := httptest.NewRecorder()

	c := e.NewContext(req, res)

	if assert.NoError(t, h(c)) {
		if assert.NoError(t, baseLoggerStatusOK(c)) {
			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, "", outputBuffer.String())
		}
	}
}

func baseLoggerStatusOK(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func baseLoggerStatusNotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, nil)
}
