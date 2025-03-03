package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	pb "github.com/duolacloud/micro/examples/grpc_with_metrics_demo/gen/go/hello"
	"github.com/duolacloud/micro/grpc/server"
	http_server "github.com/duolacloud/micro/http"
	"github.com/duolacloud/micro/logging"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/oklog/run"
)

type HelloHandler struct {
	pb.UnimplementedHelloServiceServer
}

func (h *HelloHandler) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s", req.Name),
	}, nil
}

func main() {
	l, _ := zap.NewDevelopment()
	logger := logging.NewLogger(l)

	// Setup metrics.
	metrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	registerFn := func(s *grpc.Server) {
		reflection.Register(s)
		pb.RegisterHelloServiceServer(s, &HelloHandler{})
	}

	svr := server.NewGrpcServer(
		registerFn,
		server.WithPort(50052),
		server.WithStatsHandler(otelgrpc.NewServerHandler()),
		server.WithServerMetrics(metrics),
		server.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				metrics.UnaryServerInterceptor(),
				grpc_validator.UnaryServerInterceptor(),
			),
		),
		server.WithStreamInterceptor(
			grpc_middleware.ChainStreamServer(
				metrics.StreamServerInterceptor(),
				grpc_validator.StreamServerInterceptor(),
			),
		),
	)

	g := run.Group{}
	g.Add(func() error {
		return svr.Start()
	}, func(err error) {
		svr.Stop()
	})

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	metricsSrv := http_server.NewServer(
		http_server.WithPort(8080),
		http_server.WithHandler(handler),
	)
	g.Add(func() error {
		return metricsSrv.Start()
	}, func(err error) {
		metricsSrv.Stop()
	})

	// g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	if err := g.Run(); err != nil {
		logger.Error("program interrupted", zap.Error(err))
		os.Exit(1)
	}
}
