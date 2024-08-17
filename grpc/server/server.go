package server

import (
	"fmt"
	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/stats"
)

// RegisterFn register object
type RegisterFn func(srv *grpc.Server)

// ServiceRegisterFn service register
type ServiceRegisterFn func()

// Option set server option
type Option func(*options)

type options struct {
	credentials        credentials.TransportCredentials
	statsHandler       stats.Handler
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	port               int
	serviceRegisterFn  ServiceRegisterFn
	tracingEnabled     bool
}

func defaultServerOptions() *options {
	return &options{tracingEnabled: true}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithSecure set secure
func WithSecure(credential credentials.TransportCredentials) Option {
	return func(o *options) {
		o.credentials = credential
	}
}

func WithTracingEnabled(tracingEnabled bool) Option {
	return func(o *options) {
		o.tracingEnabled = tracingEnabled
	}
}

func WithStatsHandler(statsHandler stats.Handler) Option {
	return func(o *options) {
		o.statsHandler = statsHandler
	}
}

// WithUnaryInterceptor set unary interceptor
func WithUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor) Option {
	return func(o *options) {
		o.unaryInterceptors = interceptors
	}
}

// WithStreamInterceptor set stream interceptor
func WithStreamInterceptor(interceptors ...grpc.StreamServerInterceptor) Option {
	return func(o *options) {
		o.streamInterceptors = interceptors
	}
}

// WithServiceRegister set service register
func WithServiceRegister(fn ServiceRegisterFn) Option {
	return func(o *options) {
		o.serviceRegisterFn = fn
	}
}

func customInterceptorOptions(o *options) []grpc.ServerOption {
	var opts []grpc.ServerOption

	if o.credentials != nil {
		opts = append(opts, grpc.Creds(o.credentials))
	}

	if o.tracingEnabled {
		opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	if o.statsHandler != nil {
		opts = append(opts, grpc.StatsHandler(o.statsHandler))
	}

	if len(o.unaryInterceptors) > 0 {
		option := grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(o.unaryInterceptors...),
		)
		opts = append(opts, option)
	}
	if len(o.streamInterceptors) > 0 {
		option := grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(o.streamInterceptors...),
		)
		opts = append(opts, option)
	}

	return opts
}

func WithPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

type GrpcServer struct {
	srv  *grpc.Server
	opts *options
}

func (s *GrpcServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.opts.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := s.srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func (s *GrpcServer) Stop() error {
	fmt.Println("GRPC Server stop")
	return nil
}

func (s *GrpcServer) Scheme() string {
	return "GRPC"
}

func (s *GrpcServer) Addr() string {
	return fmt.Sprintf(":%d", s.opts.port)
}

func NewGrpcServer(registerFn RegisterFn, options ...Option) *GrpcServer {
	o := defaultServerOptions()
	o.apply(options...)

	srv := grpc.NewServer(customInterceptorOptions(o)...)

	// register object to the server
	registerFn(srv)

	// register service to target
	if o.serviceRegisterFn != nil {
		o.serviceRegisterFn()
	}

	return &GrpcServer{
		srv:  srv,
		opts: o,
	}
}
