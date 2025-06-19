package logger

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Init() {
	levelStr := os.Getenv("LOG_LEVEL")
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	Log.SetLevel(level)

	if os.Getenv("LOG_FILE") == "" {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "02-Jan-2006 15:04:05",
		})
	} else {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	if filePath := os.Getenv("LOG_FILE"); filePath != "" {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			log.Printf("logger: mkdir %v", err)
		} else {
			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err == nil {
				Log.SetOutput(f)
				return
			}
			log.Printf("logger: open %v", err)
		}
	}
}
