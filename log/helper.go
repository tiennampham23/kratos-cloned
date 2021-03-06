package log

import "fmt"

var DefaultMessageKey = "msg"

type Helper struct {
	logger Logger
	msgKey string
}

type Option func(*Helper)

func NewHelper(logger Logger, opts ...Option) *Helper {
	options := &Helper{
		msgKey: DefaultMessageKey,
		logger: logger,
	}
	for _, o := range opts {
		o(options)
	}
	return options
}

func (h *Helper) Log(level Level, kvs ...interface{}) {
	_ = h.logger.Log(level, kvs...)
}

func (h *Helper) Errorf(format string, kv ...interface{}) {
	h.Log(LevelError, h.msgKey, fmt.Sprintf(format, kv...))
}

func (h *Helper) Debug(kv ...interface{}) {
	h.Log(LevelDebug, h.msgKey, fmt.Sprint(kv...))
}

// Debugw logs a message at debugg level
func (h *Helper) Debugw(kv ...interface{}) {
	h.Log(LevelDebug, kv...)
}