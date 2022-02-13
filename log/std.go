package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
)

type stdLogger struct {
	log *log.Logger
	pool *sync.Pool
}

func NewStdLogger(w io.Writer) Logger {
	return &stdLogger{
		log: log.New(w, "", 0),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (l *stdLogger) Log(level Level, kv ...interface{}) error {
	if len(kv) == 0 {
		return nil
	}
	if (len(kv) & 1) == 1 {
		kv = append(kv, "KEY_VALUES UNPAIRED")
	}
	buf := l.pool.Get().(*bytes.Buffer)
	buf.WriteString(level.String())
	for i := 0; i < len(kv); i += 2 {
		_, _ = fmt.Fprintf(buf, " %s=%v", kv[i], kv[i+1])
	}
	_ = l.log.Output(4, buf.String())
	buf.Reset()
	l.pool.Put(buf)
	return nil
}
