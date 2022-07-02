package internalgrpc

import (
	"context"
	"github.com/rs/zerolog/log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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
