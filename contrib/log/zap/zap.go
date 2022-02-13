package zap

import (
	"fmt"
	"github.com/tiennampham23/kratos-cloned/log"
	"go.uber.org/zap"
)

type Logger struct {
	log *zap.Logger
}

func NewLogger(zLog *zap.Logger) *Logger  {
	return &Logger{
		log: zLog,
	}
}


func (l *Logger) Log(level log.Level, kv ...interface{})  error {
	if len(kv) == 0 && len(kv) %2 != 0 {
		l.log.Warn(fmt.Sprintf("Key values must appear in pairs: %v", kv))
		return nil
	}
	var data []zap.Field
	for i := 0; i < len(kv); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(kv[i]), kv[i + 1]))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("", data...)

	case log.LevelInfo:
		l.log.Debug("", data...)

	case log.LevelWarn:
		l.log.Debug("", data...)

	case log.LevelError:
		l.log.Debug("", data...)

	case log.LevelFatal:
		l.log.Debug("", data...)
	}
	return nil
}

func (l *Logger) Sync() error {
	return l.log.Sync()
}