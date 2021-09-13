package config

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/color"
	"github.com/valyala/fasttemplate"
)

// The list of template tags. Used in 'middleware.DefaultLoggerConfig.Format'
// The list created according to echo framework's source code
const (
	timeUnix        templateTag = "time_unix"
	timeUnixNano    templateTag = "time_unix_nano"
	timeRfc3339     templateTag = "time_rfc3339"
	timeRfc3339Nano templateTag = "time_rfc3339_nano"
	timeCustom      templateTag = "time_custom"
	id              templateTag = "id"
	remoteIP        templateTag = "remote_ip"
	host            templateTag = "host"
	uri             templateTag = "uri"
	method          templateTag = "method"
	path            templateTag = "path"
	protocol        templateTag = "protocol"
	referer         templateTag = "referer"
	userAgent       templateTag = "user_agent"
	status          templateTag = "status"
	tagErr          templateTag = "error"
	latency         templateTag = "latency"
	latencyHuman    templateTag = "latency_human"
	bytesIn         templateTag = "bytes_in"
	bytesOut        templateTag = "bytes_out"
)

const (
	httpStatus500 = 500
	httpStatus400 = 400
	httpStatus300 = 300
)

type templateTag string

type LoggerConfig struct {
	config *middleware.LoggerConfig
	ctx    echo.Context

	template *fasttemplate.Template
	colorer  *color.Color
}

func NewLoggerConfig(config *middleware.LoggerConfig) *LoggerConfig {
	cnf := &LoggerConfig{
		config: config,
	}

	cnf.template = fasttemplate.New(cnf.config.Format, "${", "}")
	cnf.colorer = color.New()
	cnf.colorer.SetOutput(config.Output)

	return cnf
}

func (lc *LoggerConfig) DefaultLogger(startTime time.Time) ([]byte, error) {
	var buf = &bytes.Buffer{}
	defer buf.Reset()

	_, err := lc.template.ExecuteFunc(buf, lc.templateTagSwitcher(buf, startTime))

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (lc *LoggerConfig) SetContext(c echo.Context) {
	lc.ctx = c
}

func (lc *LoggerConfig) isContextValid() bool {
	return lc.ctx != nil
}

// nolint: gocyclo,funlen
func (lc *LoggerConfig) templateTagSwitcher(buf *bytes.Buffer, startTime time.Time) func(io.Writer, string) (int, error) {
	var err error

	const intFormatBase = 10

	// templateTagSwitcher function will return non-replaced log content if there is no context
	if !lc.isContextValid() {
		return func(writer io.Writer, s string) (int, error) {
			return 0, nil
		}
	}

	req := lc.ctx.Request()
	res := lc.ctx.Response()

	stop := time.Now()

	return func(w io.Writer, tag string) (int, error) {
		switch templateTag(tag) {
		case timeUnix:
			return buf.WriteString(strconv.FormatInt(time.Now().Unix(), intFormatBase))
		case timeUnixNano:
			return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), intFormatBase))
		case timeRfc3339:
			return buf.WriteString(time.Now().Format(time.RFC3339))
		case timeRfc3339Nano:
			return buf.WriteString(time.Now().Format(time.RFC3339Nano))
		case timeCustom:
			return buf.WriteString(time.Now().Format(lc.config.CustomTimeFormat))
		case id:
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = res.Header().Get(echo.HeaderXRequestID)
			}
			return buf.WriteString(requestID)
		case remoteIP:
			return buf.WriteString(lc.ctx.RealIP())
		case host:
			return buf.WriteString(req.Host)
		case uri:
			return buf.WriteString(req.RequestURI)
		case method:
			return buf.WriteString(req.Method)
		case path:
			p := req.URL.Path
			if p == "" {
				p = "/"
			}
			return buf.WriteString(p)
		case protocol:
			return buf.WriteString(req.Proto)
		case referer:
			return buf.WriteString(req.Referer())
		case userAgent:
			return buf.WriteString(req.UserAgent())
		case status:
			n := res.Status
			s := lc.colorer.Green(n)
			switch {
			case n >= httpStatus500:
				s = lc.colorer.Red(n)
			case n >= httpStatus400:
				s = lc.colorer.Yellow(n)
			case n >= httpStatus300:
				s = lc.colorer.Cyan(n)
			}
			return buf.WriteString(s)
		case tagErr:
			if err != nil {
				// Error may contain invalid JSON e.g. `"`
				b, _ := json.Marshal(err.Error())
				b = b[1 : len(b)-1]
				return buf.Write(b)
			}
		case latency:
			l := stop.Sub(startTime)
			return buf.WriteString(strconv.FormatInt(int64(l), intFormatBase))
		case latencyHuman:
			return buf.WriteString(stop.Sub(startTime).String())
		case bytesIn:
			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}
			return buf.WriteString(cl)
		case bytesOut:
			return buf.WriteString(strconv.FormatInt(res.Size, intFormatBase))
		default:
			switch {
			case strings.HasPrefix(tag, "header:"):
				return buf.Write([]byte(req.Header.Get(tag[7:])))
			case strings.HasPrefix(tag, "query:"):
				return buf.Write([]byte(lc.ctx.QueryParam(tag[6:])))
			case strings.HasPrefix(tag, "form:"):
				return buf.Write([]byte(lc.ctx.FormValue(tag[5:])))
			case strings.HasPrefix(tag, "cookie:"):
				var cookie *http.Cookie
				cookie, err = lc.ctx.Cookie(tag[7:])
				if err == nil {
					return buf.Write([]byte(cookie.Value))
				}
			}
		}
		return 0, nil
	}
}
