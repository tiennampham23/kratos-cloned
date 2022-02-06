package kratos_cloned

import (
	"context"
	"fmt"
	"github.com/tiennampham23/kratos-cloned/registry"
	"github.com/tiennampham23/kratos-cloned/transport/http"
	"sync"
	"testing"
	"time"
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


func TestApp(t *testing.T) {
	hs := http.NewServer()
	app := New(
		Name("kratos"),
		Version("v1.0.0"),
		Server(hs),
		Registrar(&mockRegistry{
			service: make(map[string]*registry.ServiceInstance),
		}),
	)
	time.AfterFunc(time.Second, func() {
		_ = app.Stop()
	})
	if err := app.Run(); err != nil {
		t.Fatal(err)
	}
}