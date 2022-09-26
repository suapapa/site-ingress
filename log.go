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
		Out:       os.Stderr,
		Level:     logrus.InfoLevel,
		Hooks:     make(logrus.LevelHooks),
		Formatter: newLogFormatter(),
	}

	hostname, _ := os.Hostname()

	log = logger.WithFields(logrus.Fields{
		"hostname": hostname,
		"program":  programName,
		"ver":      programVer,
	})
}

// log formatter to print log in KST timezone
type logFommater struct {
	ptf *prefixed.TextFormatter
	loc *time.Location
}

func newLogFormatter() *logFommater {
	ptf := prefixed.TextFormatter{
		ForceColors:     true,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceFormatting: true,
	}

	return &logFommater{
		ptf: &ptf,
		loc: time.FixedZone("KST", +9*60*60),
	}
}

func (f *logFommater) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(f.loc)
	return f.ptf.Format(e)
}
