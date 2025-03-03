package http

import (
	"fmt"
	"net/http"
)

// Option set server option
type Option func(*options)

type options struct {
	handler http.Handler
	port    int
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

func WithHandler(handler http.Handler) Option {
	return func(o *options) {
		o.handler = handler
	}
}

func defaultServerOptions() *options {
	return &options{
		port: 8080,
	}
}

type Server struct {
	opts *options
}

func (s *Server) Start() error {
	fmt.Printf("Http server start..., port: %v\n", s.opts.port)

	return http.ListenAndServe(fmt.Sprintf(":%v", s.opts.port), s.opts.handler)
}

func (s *Server) Stop() error {
	fmt.Println("Http server stop")
	return nil
}

func NewServer(options ...Option) *Server {
	o := defaultServerOptions()
	o.apply(options...)

	return &Server{
		o,
	}
}
