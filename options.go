package kratos_cloned

import (
	"context"
	"github.com/tiennampham23/kratos-cloned/registry"
	"github.com/tiennampham23/kratos-cloned/transport"
	"net/url"
	"os"
	"time"
)

// Option is an application option
type Option func(o *options)

// options is an applications options.
type options struct {
	id        string
	name      string
	version   string
	metadata  map[string]string
	endpoints []*url.URL

	ctx  context.Context
	sigs []os.Signal

	registrarTimeout time.Duration
	stopTimeout      time.Duration

	registrar registry.Registrar

	servers []transport.Server
}

func ID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

func Name(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

func Version(version string) Option {
	return func(o *options) {
		o.version = version
	}
}
func Metadata(m map[string]string) Option {
	return func(o *options) {
		o.metadata = m
	}
}

func Server(srv ...transport.Server) Option {
	return func(o *options) {
		o.servers = srv
	}
}

func Registrar(r registry.Registrar) Option {
	return func(o *options) {
		o.registrar = r
	}
}
