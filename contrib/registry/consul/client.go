package consul

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/tiennampham23/kratos-cloned/log"
	"github.com/tiennampham23/kratos-cloned/registry"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ServiceResolver is used to resolve service endpoints
type ServiceResolver func(ctx context.Context, entries []*api.ServiceEntry) []*registry.ServiceInstance
type Client struct {
	cli                 *api.Client
	ctx                 context.Context
	cancel              context.CancelFunc
	resolver            ServiceResolver
	healthCheckInterval int
	heartBeat           bool
}

func NewClient(cli *api.Client) *Client{
	c := &Client{
		cli: cli,
		resolver: defaultResolver,
		healthCheckInterval: 10,
		heartBeat: true,
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return c
}

func defaultResolver(_ context.Context, entries []*api.ServiceEntry) []*registry.ServiceInstance {
	services := make([]*registry.ServiceInstance, 0, len(entries))
	for _, entry := range entries {
		var version string
		for _, tag := range entry.Service.Tags {
			ss := strings.SplitN(tag, "=", 2)
			if len(ss) == 2 && ss[0] == "version" {
				version = ss[1]
			}
		}
		var endpoints []string
		for schema, addr := range entry.Service.TaggedAddresses {
			if schema == "lan_ipv4" || schema == "wan_ipv4" || schema == "lan_ipv6" || schema == "wan_ipv6" {
				continue
			}
			endpoints = append(endpoints, addr.Address)
		}
		services = append(services, &registry.ServiceInstance{
			ID:        entry.Service.ID,
			Name:      entry.Service.Service,
			Metadata:  entry.Service.Meta,
			Version:   version,
			Endpoints: endpoints,
		})
	}
	return services

}
func (c *Client) Register(_ context.Context, svc *registry.ServiceInstance, enableHealthCheck bool) error {
	addresses := make(map[string]api.ServiceAddress)
	checkAddresses := make([]string, 0, len(svc.Endpoints))
	for _, endpoint := range svc.Endpoints {
		raw, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		addr := raw.Hostname()
		port, err := strconv.ParseUint(raw.Port(), 10, 16)
		checkAddresses = append(checkAddresses, fmt.Sprintf("%s:%d", addr, port))
		addresses[raw.Scheme] = api.ServiceAddress{
			Address: endpoint,
			Port: int(port),
		}
	}
	asr := &api.AgentServiceRegistration{
		ID: svc.ID,
		Name: svc.Name,
		Meta: svc.Metadata,
		Tags: []string{fmt.Sprintf("version=%s", svc.Version)},
		TaggedAddresses: addresses,
	}
	if len(checkAddresses) > 0 {
		host, portRaw, _ := net.SplitHostPort(checkAddresses[0])
		port, _ := strconv.ParseInt(portRaw, 10, 32)
		asr.Address = host
		asr.Port = int(port)
	}
	if enableHealthCheck {
		for _, address := range checkAddresses {
			asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
				TCP: address,
				Interval: fmt.Sprintf("%ds", c.healthCheckInterval),
				DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.healthCheckInterval*60),
				Timeout: "5s",
			})
		}
	}
	if c.heartBeat {
		asr.Checks = append(asr.Checks, &api.AgentServiceCheck{
			CheckID: "service:" + svc.ID,
			TTL: fmt.Sprintf("%ds", c.healthCheckInterval*2),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", c.healthCheckInterval*60),
		})
	}
	err := c.cli.Agent().ServiceRegister(asr)
	if err != nil {
		return err
	}
	if c.heartBeat {
		go func() {
			time.Sleep(time.Second)
			err := c.cli.Agent().UpdateTTL("service:" + svc.ID, "pass", "pass")
			if err != nil {
				log.Errorf("[Consul] Update TTL heartbeat to consul failed with: %v", err)
			}
			ticker := time.NewTicker(time.Second * time.Duration(c.healthCheckInterval))
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					err := c.cli.Agent().UpdateTTL("service:" + svc.ID, "pass", "pass")
					if err != nil {
						log.Errorf("[Consul] Update TTL heartbeat to consul failed with: %v", err)
					}
				case <- c.ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

func (c *Client) Deregister(ctx context.Context, serviceId string) error {
	c.cancel()
	return c.cli.Agent().ServiceDeregister(serviceId)
}

func (c *Client) Service(ctx context.Context, serviceName string, index uint64, passingOnly bool) ([]*registry.ServiceInstance, uint64, error) {
	opts := &api.QueryOptions{
		WaitIndex: index,
		WaitTime: time.Second * 55,
	}
	opts = opts.WithContext(ctx)
	entries, meta, err := c.cli.Health().Service(serviceName, "", passingOnly, opts)
	if err != nil {
		return nil, 0, err
	}
	return c.resolver(ctx, entries), meta.LastIndex, nil
}