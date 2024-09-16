package loggers

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"time"
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

// CustomFormatter для кастомизации формата логов
type CustomFormatter struct{}

// Format задает формат лога
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Получаем информацию о файле и строке
	_, file, line, ok := runtime.Caller(8) // 8 уровней вверх по стеку, чтобы получить нужный файл и строку
	if !ok {
		file = "unknown"
		line = 0
	}

	// Выделяем только имя файла из полного пути
	fileName := file
	if lastSlash := len(file) - 1; lastSlash >= 0 {
		for i := len(file) - 1; i >= 0; i-- {
			if file[i] == '/' || file[i] == '\\' {
				fileName = file[i+1:]
				break
			}
		}
	}

	// Формируем строку лога
	log := fmt.Sprintf("%s [%s] %s:%d %s\n",
		time.Now().Format(time.RFC3339), // Время
		entry.Level.String(),            // Уровень лога
		fileName, line,                  // Имя файла и строка
		entry.Message, // Сообщение
	)

	return []byte(log), nil
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
	//_, file, line, _ := runtime.Caller(2)
	//location := fmt.Sprintf("%s:%d", file, line)

	if l.level >= level {
		l.log.Log(level, v...)

		/*l.log.WithFields(logrus.Fields{
			"location": location,
		}).Log(level, v...)*/
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

func CreateLogger(cfg LoggerConfig) Logger {
	var invalidCfg bool
	if cfg.valid() == false {
		invalidCfg = true
		cfg.Name = "Default"
	}

	for _, logger := range loggers {
		if cfg.Name == logger.GetName() {
			return logger
		}
	}
	if invalidCfg {
		return createDefaultLogger()
	}

	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return createDefaultLogger()
	}

	var writer io.Writer
	if cfg.WriteToConsole {
		writer = io.MultiWriter(os.Stdout, file)
	} else {
		writer = file
	}

	lrus := logrus.New()

	/*lrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   cfg.UseColor,
	})*/
	lrus.SetFormatter(&CustomFormatter{})
	lrus.SetOutput(writer)

	lg := logger{
		name:  cfg.Name,
		path:  cfg.Path,
		level: logrus.Level(cfg.Level),
		log:   lrus,
	}

	loggers = append(loggers, &lg)

	return &lg
}

func createDefaultLogger() Logger {
	file, _ := os.OpenFile("./DefaultLogs.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	lrus := logrus.New()
	lrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	})
	lrus.SetOutput(file)

	lg := logger{
		name:  "Default",
		path:  "./DefaultLogs.txt",
		level: logrus.Level(InfoLevel),
		log:   lrus,
	}

	loggers = append(loggers, &lg)

	return &lg
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
