package main

import (
	"context"
	"fmt"
	kratos_cloned "github.com/tiennampham23/kratos-cloned"
	"github.com/tiennampham23/kratos-cloned/registry"
	"github.com/tiennampham23/kratos-cloned/transport/http"
	"sync"
)

type mockRegistry struct {
	lk sync.Mutex
	service map[string]*registry.ServiceInstance
}

func (r *mockRegistry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	if service == nil || service.ID == "" {
		return fmt.Errorf("no service id")
	}
	r.lk.Lock()
	defer r.lk.Unlock()
	r.service[service.ID] = service
	return nil
}

func (r *mockRegistry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	r.lk.Lock()
	defer r.lk.Unlock()
	if r.service[service.ID] == nil {
		return fmt.Errorf("deregister service not found")
	}
	delete(r.service, service.ID)
	return nil
}

func main()  {
	hs := http.NewServer()
	app := kratos_cloned.New(
		kratos_cloned.Name("kratos"),
		kratos_cloned.Version("v1.0.0"),
		kratos_cloned.Server(hs),
		kratos_cloned.Registrar(&mockRegistry{
			service: make(map[string]*registry.ServiceInstance),
		}),
	)
	if err := app.Run(); err != nil {
		fmt.Println(err.Error())
	}
}
