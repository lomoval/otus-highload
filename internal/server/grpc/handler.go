package internalgrpc

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var (
	errorRate = promauto.NewCounter(prometheus.CounterOpts{
		Name: "otus_hl_dialogs_error_total",
	})
	requestRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otus_hl_dialogs_req_total",
		},
		[]string{"method"},
	)
	latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "otus_hl_dialogs_latency",
		},
		[]string{"method"},
	)
)

func loggingHandler(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("method %q failed: %s", info.FullMethod, err)
	}
	ip := ""
	if peer, ok := peer.FromContext(ctx); ok {
		ip = peer.Addr.String()
	}
	var userAgent []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userAgent = md.Get("user-agent")
	}
	log.Info().
		Str("ip", ip).
		Str("method", info.FullMethod).
		Strs("user-agent", userAgent).
		Str("latency", time.Since(start).String()).
		Msg("GRPC request processed")
	return resp, err
}

func serverLoggingMetricsInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	requestRate.WithLabelValues(info.FullMethod).Inc()

	h, err := handler(ctx, req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("method %q failed: %s", info.FullMethod, err)
		errorRate.Inc()
	}
	ip := ""
	if peer, ok := peer.FromContext(ctx); ok {
		ip = peer.Addr.String()
	}
	var userAgent []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userAgent = md.Get("user-agent")
	}
	latency.WithLabelValues(info.FullMethod).Observe(float64(duration.Milliseconds()))
	log.Info().
		Str("ip", ip).
		Str("method", info.FullMethod).
		Strs("user-agent", userAgent).
		Str("latency", duration.String()).
		Msg("GRPC request processed")
	return h, err
}
