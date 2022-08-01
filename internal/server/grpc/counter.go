package internalgrpc

import (
	"app/api"
	"app/service"
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
	"time"
)

type Counter struct {
	grpcServer *grpc.Server
	addr       string
	api.UnimplementedCounterServer
}

func NewCounterServer(host string, port int) *Counter {
	return &Counter{addr: net.JoinHostPort(host, strconv.Itoa(port))}
}

func (s *Counter) Start(_ context.Context) error {
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(loggingHandler))
	api.RegisterCounterServer(s.grpcServer, s)

	lsn, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Err(err).Msgf("failed to listen grpc endpoint: %v", err)
		return err
	}

	log.Printf("starting grpc server on %s", s.addr)
	err = s.grpcServer.Serve(lsn)
	return err
}

func (s *Counter) Stop() {
	s.grpcServer.Stop()
}

func (s *Counter) PrivateDialogsCounters(_ context.Context, request *api.PrivateDialogsCountersRequest) (*api.PrivateDialogsCountersResponse, error) {
	dialogs, err := service.PrivateDialogCounters(request.UserId)
	return &api.PrivateDialogsCountersResponse{Dialogs: dialogs}, err
}

func (s *Counter) ResetPrivateDialogCounter(_ context.Context, request *api.ResetPrivateDialogCounterRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, service.ResetPrivateDialogCounter(request.DialogId, request.OwnerId, time.Now().UTC())
}
