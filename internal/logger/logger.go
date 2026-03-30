package logger

import (
	"os"

	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, err
	}

	config := zap.NewProductionConfig()

	config.OutputPaths = []string{
		"stdout",
		"logs/app.log",
	}

	return config.Build()
}
