package zap

import (
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


func (l *Logger) Log(level log.Level)