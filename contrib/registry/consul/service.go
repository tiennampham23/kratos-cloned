package consul

import (
	"github.com/tiennampham23/kratos-cloned/registry"
	"sync"
	"sync/atomic"
)

type serviceSet struct {
	serviceName string
	lock        sync.RWMutex
	services    *atomic.Value
	watcher     map[*watcher]struct{}
}

func (s *serviceSet) broadcast(services []*registry.ServiceInstance) {
	s.services.Store(services)
	s.lock.Lock()
	defer s.lock.Unlock()
	for k := range s.watcher {
		select {
		case k.event <- struct{}{}:
		default:
		}
	}

}