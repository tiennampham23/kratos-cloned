package kratos_cloned

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tiennampham23/kratos-cloned/registry"
	"github.com/tiennampham23/kratos-cloned/transport"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

type App struct {
	opts     options
	ctx      context.Context
	cancel   func()
	lk       sync.Mutex
	instance *registry.ServiceInstance
}

type appKey struct{}

func (a *App)Stop() error {
	fmt.Println("Stopping....")
	return nil
}

// Run executes all OnStart hooks registered with the application's Lifecycle.
func (a *App) Run() error {
	fmt.Println("Running....")
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}
	ctx := NewContext(a.ctx, a)
	eg, ctx := errgroup.WithContext(ctx)
	wg := sync.WaitGroup{}
	for _, srv := range a.opts.servers {
		srv := srv
		eg.Go(func() error {
			<- ctx.Done() // wait for stop signal
			sCtx, sCancel := context.WithTimeout(NewContext(context.Background(), a), a.opts.stopTimeout)
			defer sCancel()
			return srv.Stop(sCtx)
		})
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			return srv.Start(ctx)
		})
	}
	wg.Wait()
	if a.opts.registrar != nil {
		rCtx, rCancel := context.WithTimeout(a.opts.ctx, a.opts.registrarTimeout)
		defer rCancel()
		if err := a.opts.registrar.Register(rCtx, instance); err != nil {
			return err
		}
		a.lk.Lock()
		a.instance = instance
		a.lk.Unlock()
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	eg.Go(func() error {
		for {
			select {
			case <- ctx.Done():
				return ctx.Err()
			case <- c:
				if err := a.Stop(); err != nil {
					return err
				}
			}

		}
	})
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func NewContext(ctx context.Context, a AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, a)
}

func New(opts ...Option) *App {
	o := options{
		ctx:              context.Background(),
		sigs:             []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registrarTimeout: 10 * time.Second,
		stopTimeout:      10 * time.Second,
	}
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}

	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   o,
	}
}


func (a *App) ID() string {
	return a.opts.id
}

func (a *App) Name() string {
	return a.opts.name
}

func (a *App) Version() string {
	return a.opts.version
}

func (a *App) Metadata() map[string]string {
	return a.opts.metadata
}

func (a *App) Endpoint() []string {
	if a.instance == nil {
		return []string{}
	}
	return a.instance.Endpoints
}

func (a *App) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string,  0)
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}
	if len(endpoints) == 0 {
		for _, srv := range a.opts.servers {
			if r, ok := srv.(transport.Enpointer); ok {
				e, err := r.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, e.String())
			}
		}
	}
	return &registry.ServiceInstance{
		ID: a.opts.id,
		Name: a.opts.name,
		Version: a.opts.version,
		Metadata: a.opts.metadata,
		Endpoints: endpoints,
	}, nil
}