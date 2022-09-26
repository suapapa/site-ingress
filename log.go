package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	log *logrus.Entry
)

func init() {
	initLogger()
}

func initLogger() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.InfoLevel,
		Hooks: make(logrus.LevelHooks),
		Formatter: &prefixed.TextFormatter{
			ForceColors:     true,
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}

	hostname, _ := os.Hostname()

	log = logger.WithFields(logrus.Fields{
		"hostname": hostname,
		"program":  programName,
		"ver":      programVer,
	})
}
