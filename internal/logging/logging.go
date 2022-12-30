package logging

import (
	"fmt"
	"os"

	"mi-ca/internal/env"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type Logger = zerolog.Logger

var (
	Log Logger
)

var CallerPrettyfierFunc = func(pc uintptr, file string, line int) string {
	var short string
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}
	return fmt.Sprintf("%s:%d", short, line)
}

func Init() {
	switch env.APP_MODE {
	case "DEV":
		Log = zlog.With().Caller().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr,
			TimeFormat: "2006-01-02 15:04:05.000000", FormatCaller: func(i interface{}) string {
				if s, ok := i.(string); ok {
					for i := len(s) - 1; i > 0; i-- {
						if s[i] == '/' {
							return s[i+1:]
						}
					}
				}

				return fmt.Sprintf("%s", i)
			}})
		zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000000"
	default:
		zerolog.CallerFieldName = "caller"
		zerolog.CallerMarshalFunc = CallerPrettyfierFunc
		zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000000"
		Log = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	}

	switch env.LOG_LEVEL {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		fmt.Printf("Log level %s unsupported, defaulting to info\n", env.LOG_LEVEL)
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
