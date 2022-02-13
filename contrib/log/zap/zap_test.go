package zap

import (
	"github.com/tiennampham23/kratos-cloned/log"
	"go.uber.org/zap"
	"testing"
)

func Test_ZapLogger(t *testing.T) {
	logger := NewLogger(zap.NewExample())
	defer func() {
		_ = logger.Sync()
	}()
	zLogger := log.NewHelper(logger)
	zLogger.Debugw("log", "debug")
}