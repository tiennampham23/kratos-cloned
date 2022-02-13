package log

import "sync"

// globalLogger is designed as a global logger in current process.
var global = &loggerAppliance{}

// loggerAppliance is the proxy of `Logger` to make logger change will affect to all sub-logger
type loggerAppliance struct {
	lock sync.Mutex
	Logger
	helper *Helper
}

func init() {
	global.SetLogger(DefaultLogger)
}

// SetLogger should be called before any other log call.
// And it is NOT THREAD SAFE
func SetLogger(logger Logger) {
	global.SetLogger(logger)
}

func (a *loggerAppliance) SetLogger(in Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
	a.helper = NewHelper(a.Logger)
}

// Errorf logs a message at error level.
func Errorf(format string, kv ...interface{}) {
	global.helper.Errorf(format, kv...)
}
