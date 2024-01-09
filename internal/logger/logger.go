package logger

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Options struct {
	Level     string
	AddSource bool
}

func init() {
	Set(Options{})
}

func Set(options Options) {
	if options.Level == "" {
		options.Level = "INFO"
	}

	var sLevel slog.Level

	switch options.Level {
	case "WARNING":
		sLevel = slog.LevelWarn
	case "INFO":
		sLevel = slog.LevelInfo
	case "DEBUG":
		sLevel = slog.LevelDebug
	default:
		sLevel = slog.LevelError
	}

	programLevel := new(slog.LevelVar)
	programLevel.Set(sLevel)

	opts := slog.HandlerOptions{
		Level:     programLevel,
		AddSource: options.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// Parse it correctly
				a.Key = "date" // Rename time into date
				a.Value = slog.AnyValue(time.Now().Truncate(time.Millisecond))
			}
			return a
		},
	}

	h := slog.NewJSONHandler(os.Stdout, &opts)
	slog.SetDefault(slog.New(h))

	slog.Info(fmt.Sprintf("slog level %s", sLevel))
}
