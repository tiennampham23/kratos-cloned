package consul

import (
	"context"
	"github.com/hashicorp/consul/api"
	"github.com/tiennampham23/kratos-cloned/registry"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	*api.Config
}
type Registry struct {
	enableHealthCheck bool
	lock              sync.RWMutex
	cli               *Client
	registry          map[string]*serviceSet
}
type Option func(*Registry)

func New(apiClient *api.Client, opts ...Option) *Registry {
	r := &Registry{
		cli:               NewClient(apiClient),
		registry:          make(map[string]*serviceSet),
		enableHealthCheck: true,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}
// WithHeartbeat enable or disable heartbeat
func WithHeartbeat(enable bool) Option {
	return func(r *Registry) {
		if r.cli != nil {
			r.cli.heartBeat = enable
		}
	}
}

// WithHealthCheck with registry health check option.
func WithHealthCheck(enable bool) Option {
	return func(r *Registry) {
		r.enableHealthCheck = enable
	}
}

// WithHealthCheckInterval with health check interval in seconds.
func WithHealthCheckInterval(interval int) Option {
	return func(r *Registry) {
		if r.cli != nil {
			r.cli.healthCheckInterval = interval
		}
	}
}

func (r *Registry) Register(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Register(ctx, svc, r.enableHealthCheck)
}

func (r *Registry) Deregister(ctx context.Context, svc *registry.ServiceInstance) error {
	return r.cli.Deregister(ctx, svc.ID)
}


// Watch resolve service by name
func (r *Registry) Watch(ctx context.Context, name string)  (registry.Watcher, error){
	r.lock.Lock()
	defer r.lock.Unlock()
	set, ok := r.registry[name]
	if !ok {
		set = &serviceSet{
			watcher: make(map[*watcher]struct{}),
			services: &atomic.Value{},
			serviceName: name,
		}
	}
	w := &watcher{
		event: make(chan struct{}, 1),
	}
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.set = set
	set.lock.Lock()
	set.watcher[w] = struct{}{}
	set.lock.Unlock()
	ss, _ := set.services.Load().([]*registry.ServiceInstance)
	if len(ss) > 0 {
		// If the service has a value, it needs to be pushed to the watcher,
		// otherwise the initial data may be blocked forever during the watch.
		w.event <- struct{}{}
	}
	if !ok {
		err := r.resolve(set)
		if err != nil {
			return nil, err
		}
	}
	return w, nil
}

func (r *Registry) resolve(set *serviceSet) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	services, idx, err := r.cli.Service(ctx, set.serviceName, 0, true)
	cancel()
	if err != nil {
		return nil
	} else if len(services) > 0 {
		set.broadcast(services)
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			<- ticker.C
			ctx, cancel := context.WithTimeout(context.Background(), time.Second * 120)
			tmpService, tmpIdx, err := r.cli.Service(ctx, set.serviceName, idx, true)
			cancel()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			if len(tmpService) != 0 && tmpIdx != idx {
				services = tmpService
				set.broadcast(services)
			}
			idx = tmpIdx
		}
	}()
	return nil
}
