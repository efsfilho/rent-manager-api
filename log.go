package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var customLogger zerolog.Logger

type state struct {
	b []byte
}

func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

func (s *state) Flag(c int) bool {
	return false
}

func frameField(f errors.Frame, s *state, c rune) string {
	f.Format(s, c)
	return string(s.b)
}

// Reimplementation of pkgerrors.MarshalStack with a different stacktrace format
func MarshalStackTest(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	sterr, ok := err.(stackTracer)
	if !ok {
		return nil
	}
	st := sterr.StackTrace()
	s := &state{}

	var out strings.Builder
	out.WriteString("\n")
	for i := range st {
		file := frameField(st[len(st)-1-i], s, 's')
		line := frameField(st[len(st)-1-i], s, 'd')
		fun := frameField(st[len(st)-1-i], s, 'n')
		out.WriteString(fmt.Sprintf("\t%-35s %v:%v \n", fun, file, line))
	}
	return out.String()
}

func setLog(level zerolog.Level) {
	// Should be used with errors.Wrap() from "github.com/pkg/errors"
	// zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.ErrorStackMarshaler = MarshalStackTest

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(level)

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Logging level
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("[ %-5s ]", i))
	}

	// Allow special chars when logging stacktrace errors
	output.FormatFieldValue = func(i interface{}) string {
		value, ok := i.(string)
		if ok {
			i, _ = strconv.Unquote(value)
		}
		return fmt.Sprintf("%s ", i)
	}

	// Sets a custom logger to be used by logger()
	customLogger = zerolog.New(output).With().Timestamp().Logger()
}

// logger function globally used
func logger(log interface{}) {
	// Error handling
	errorValue, isError := log.(error)
	if isError {
		wraped := errors.Wrap(errorValue, "")
		customLogger.Error().Stack().Err(wraped).Msg("")
		return
	}

	// Simple log
	msgValue, isString := log.(string)
	if isString {
		customLogger.Info().Msg(msgValue)
		return
	}
}

// Requests log format config
var RequestLoggerConfig = middleware.RequestLoggerConfig{
	LogMethod:       true,
	LogURI:          true,
	LogProtocol:     true,
	LogStatus:       true,
	LogResponseSize: true,
	LogRemoteIP:     true,
	LogUserAgent:    true,
	LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
		msg := fmt.Sprintf("%-6s %v %v %v %v %v ",
			v.Method,
			v.Status,
			v.URI,
			v.ResponseSize,
			v.Protocol,
			v.UserAgent)
		customLogger.Info().Msg(msg)
		return nil
	},
}
