package logger

import (
	"io"
	"log/slog"
	"os"
)

func New(level string, format string, filePath string) (*slog.Logger, error) {

	//  Parse Log Level 
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	//  Open Log File 
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	//  Multi Writer (Console + File) 
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	//  Select Format 
	var handler slog.Handler

	if format == "json" {
		opts := &slog.HandlerOptions{
			Level:     slogLevel,
			AddSource: true,
		}
		handler = slog.NewJSONHandler(multiWriter, opts)
	} else {
		handler = NewTextHandler(multiWriter, slogLevel)
	}

	return slog.New(handler), nil
}
