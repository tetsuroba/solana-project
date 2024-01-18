package services

import (
	"log/slog"
	"os"
)

var logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}).WithAttrs([]slog.Attr{slog.String("service", "services")})

var logger = slog.New(logHandler)
