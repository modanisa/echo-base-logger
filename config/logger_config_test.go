package config

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

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

func TestNewLoggerConfig(t *testing.T) {
	_init()
	conf := NewLoggerConfig(&middleware.DefaultLoggerConfig)

	assert.NotNil(t, conf)
}

func TestShouldReturnNonReplacedLogContentIfContextNotSet_DefaultLogger(t *testing.T) {
	_init()
	conf := NewLoggerConfig(&middleware.DefaultLoggerConfig)

	logContent, err := conf.DefaultLogger(time.Now())

	expectedLogContent := `{"time":"","id":"","remote_ip":"","host":"","method":"","uri":"","user_agent":"","status":,"error":"","latency":,"latency_human":"","environment":"","bytes_in":,"bytes_out":}`

	assert.NoError(t, err)
	assert.Contains(t, string(logContent), expectedLogContent)
}

func TestLoggerConfig_DefaultLogger(t *testing.T) {
	_init()
	e := echo.New()

	req := httptest.NewRequest(echo.GET, "/context-has-been-set", nil)
	res := httptest.NewRecorder()

	c := e.NewContext(req, res)

	conf := NewLoggerConfig(&middleware.DefaultLoggerConfig)
	conf.SetContext(c)

	expectedLogContent := `"host":"example.com","method":"GET","uri":"/context-has-been-set"`

	logContent, err := conf.DefaultLogger(time.Now())

	assert.NoError(t, err)
	assert.Contains(t, string(logContent), expectedLogContent)
}
