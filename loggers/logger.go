package loggers

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
)

var loggers []*logger

type Level logrus.Level

const (
	DebugLevel = Level(logrus.DebugLevel)
	InfoLevel  = Level(logrus.InfoLevel)
	WarnLevel  = Level(logrus.WarnLevel)
	ErrorLevel = Level(logrus.ErrorLevel)
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})

	GetName() string
}

type logger struct {
	name  string
	level logrus.Level
	path  string
	log   *logrus.Logger
}

type LoggerConfig struct {
	Name           string
	Path           string
	Level          Level
	WriteToConsole bool
	UseColor       bool
}

func (lc *LoggerConfig) valid() bool {
	if lc.Level != InfoLevel &&
		lc.Level != WarnLevel &&
		lc.Level != ErrorLevel &&
		lc.Level != DebugLevel {
		return false
	}
	if lc.Path == "" {
		return false
	}
	if lc.Name == "" {
		return false
	}

	return true
}

func (l *logger) logMessage(level logrus.Level, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	location := fmt.Sprintf("%s:%d", file, line)

	if l.level >= level {
		l.log.WithFields(logrus.Fields{
			"location": location,
		}).Log(level, v...)
	}
}

func (l *logger) Debug(v ...interface{}) {
	l.logMessage(logrus.Level(DebugLevel), v)
}

func (l *logger) Info(v ...interface{}) {
	l.logMessage(logrus.Level(InfoLevel), v)
}

func (l *logger) Warn(v ...interface{}) {
	l.logMessage(logrus.Level(WarnLevel), v)
}

func (l *logger) Error(v ...interface{}) {
	l.logMessage(logrus.Level(ErrorLevel), v)
}

func CreateLogger(cfg LoggerConfig) (Logger, error) {
	if cfg.valid() == false {
		return nil, errors.New("logger name is empty")
	}

	for _, logger := range loggers {
		if cfg.Name == logger.GetName() {
			return logger, nil
		}
	}

	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.New("open file fail")
	}

	var writer io.Writer
	if cfg.WriteToConsole {
		writer = io.MultiWriter(os.Stdout, file)
	} else {
		writer = file
	}

	lrus := logrus.New()
	lrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   cfg.UseColor,
	})
	lrus.SetOutput(writer)

	lg := logger{
		name:  cfg.Name,
		path:  cfg.Path,
		level: logrus.Level(cfg.Level),
		log:   lrus,
	}

	loggers = append(loggers, &lg)

	return &lg, nil
}

func (l *logger) GetName() string {
	return l.name
}

func (l *logger) SetLevel(lvl Level) error {
	switch lvl {
	case DebugLevel:
		l.level = logrus.Level(DebugLevel)

	case InfoLevel:
		l.level = logrus.Level(InfoLevel)

	case WarnLevel:
		l.level = logrus.Level(WarnLevel)

	case ErrorLevel:
		l.level = logrus.Level(ErrorLevel)

	default:
		l.level = logrus.Level(WarnLevel)
		return errors.New("invalid level")
	}

	return nil
}
