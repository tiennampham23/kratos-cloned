package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Server is an HTTP server wrapper.
type Server struct {
	*http.Server
	lis net.Listener
	tlsConf *tls.Config
	endpoint *url.URL
	err error
	network string
	address string
	timeout time.Duration
	strictSlash bool
	router *mux.Router
}

// ServerOption is an HTTP server option.
type ServerOption func(*Server)

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		timeout: 1 * time.Second,
		strictSlash: true,
	}
	for _, o := range opts {
		o(srv)
	}
	srv.router = mux.NewRouter().StrictSlash(srv.strictSlash)
	srv.Server = &http.Server{
		TLSConfig: srv.tlsConf,
	}
	srv.err = srv.listenAndEndpoint()
	return srv
}

func (s *Server) Start(ctx context.Context) error {
	if s.err != nil {
		return s.err
	}
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	fmt.Println(fmt.Sprintf("[HTTP] server listening on: %s", s.lis.Addr().String()))
	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.lis, "", "")
	} else {
		err = s.Serve(s.lis)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}

func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	return nil
}