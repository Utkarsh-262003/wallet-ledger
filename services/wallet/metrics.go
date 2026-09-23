package main

import (
	"context"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	grpcRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wallet_grpc_requests_total",
		Help: "gRPC calls by method and status code.",
	}, []string{"method", "code"})

	grpcDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "wallet_grpc_duration_seconds",
		Help:    "gRPC call duration.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})
)

// metricsInterceptor runs around every gRPC call.
// It times the call and counts it by method and result code.
func metricsInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	method := path.Base(info.FullMethod) // "/wallet.v1.WalletService/Transfer" -> "Transfer"
	grpcRequests.WithLabelValues(method, status.Code(err).String()).Inc()
	grpcDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	return resp, err
}