package consul

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/tiennampham23/kratos-cloned/registry"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	// get intranet ip
	// the intranet ip is often private ip
	addr := fmt.Sprintf("%s:9091", getIntranetIP())
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed!", addr)
		t.Fail()
	}
	defer func(lis net.Listener) {
		err := lis.Close()
		if err != nil {
			t.Fail()
		}
	}(lis)
	go tcpServer(lis)
	time.Sleep(time.Millisecond * 1000)
	cli, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
	})
	if err != nil {
		t.Fatalf("create consul client failed: %v", err.Error())
	}
	opts := []Option {
		WithHeartbeat(true),
		WithHealthCheck(true),
		WithHealthCheckInterval(5),
	}
	r := New(cli, opts...)
	version := strconv.FormatInt(time.Now().Unix(), 10)
	svc := &registry.ServiceInstance{
		ID: "test123",
		Name: "test-provider",
		Version: version,
		Metadata: map[string]string{"app": "kratos-cloned"},
		Endpoints: []string{fmt.Sprintf("tcp://%v?isSecure=false", addr)},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second* 5)
	defer cancel()
	err = r.Deregister(ctx, svc)
	if err != nil {
		t.Errorf("Degister failed, %v", err.Error())
	}
	err = r.Register(ctx, svc)

	if err != nil {
		t.Errorf("Register failed, %v", err.Error())
	}
	w, err := r.Watch(ctx, "test-provider")
	if err != nil {
		t.Errorf("Watch failed %v", err.Error())
	}
	services, err := w.Next()
	if err != nil {
		t.Errorf("Next failed %v", err.Error())
	}

	if !reflect.DeepEqual(1, len(services)) {
		t.Errorf("no expect float_key value: %v, but got: %v", len(services), 1)
	}
	if !reflect.DeepEqual("test123", services[0].ID) {
		t.Errorf("no expect float_key value: %v, but got: %v", services[0].ID, "test123")
	}
	if !reflect.DeepEqual("test-provider", services[0].Name) {
		t.Errorf("no expect float_key value: %v, but got: %v", services[0].Name, "test-provider")
	}
	if !reflect.DeepEqual(version, services[0].Version) {
		t.Errorf("no expect float_key value: %v, but got: %v", services[0].Version, version)
	}

}

func tcpServer(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}
		fmt.Println("Get tcp")
		err = conn.Close()
		if err != nil {
			return
		}
	}
}

func getIntranetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil{
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}