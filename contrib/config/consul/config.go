package consul

import (
	"context"
	"errors"
	"github.com/hashicorp/consul/api"
	"github.com/tiennampham23/kratos-cloned/config"
	"path/filepath"
	"strings"
)

type Option func(o *options)

type options struct {
	ctx context.Context
	path string
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithPath(path string) Option {
	return func(o *options) {
		o.path = path
	}
}

type source struct {
	client *api.Client
	options *options
}

func New(client *api.Client, opts ...Option) (config.Source, error) {
	options := &options{
		ctx: context.Background(),
		path: "",
	}
	for _, opt := range opts {
		opt(options)
	}
	if options.path == "" {
		return nil, errors.New("path invalid")
	}
	return &source{
		client: client,
		options: options,
	}, nil
}


func (s *source) Load() ([]*config.KeyValue, error) {
	kv, _, err := s.client.KV().List(s.options.path, nil)
	if err != nil {
		return nil, err
	}
	pathPrefix := s.options.path
	if !strings.HasSuffix(s.options.path, "/") {
		pathPrefix = pathPrefix + "/"
	}
	kvs := make([]*config.KeyValue, 0)
	for _, item := range kv {
		k := strings.TrimPrefix(item.Key, pathPrefix)
		kvs = append(kvs, &config.KeyValue{
			Key: k,
			Value: item.Value,
			Format: strings.TrimPrefix(filepath.Ext(k), "."),
		})
	}
	return kvs, nil
}

func (s *source) Watch() (config.Watcher, error) {
	return newWatcher(s)
}