package service

import (
	"app/api"
	"context"
	"github.com/simplesurance/grpcconsulresolver/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

var (
	grpcCounterConn   *grpc.ClientConn
	grpcCounterClient api.CounterClient
)

func SetupGrpcCounters(uri string) error {
	resolver.Register(consul.NewBuilder())
	var err error

	grpcCounterConn, err = grpc.Dial(
		uri,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	grpcCounterClient = api.NewCounterClient(grpcCounterConn)
	return nil
}

func StopGrpcCounter() {
	if grpcCounterConn != nil {
		grpcCounterConn.Close()
	}
}

func counterClient() api.CounterClient {
	return grpcCounterClient
}

func DialogsCounters(ctx context.Context, userID int64) (map[int64]int32, error) {
	resp, err := counterClient().PrivateDialogsCounters(ctx, &api.PrivateDialogsCountersRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return resp.Dialogs, nil
}

func ResetDialogCounter(ctx context.Context, dialogID int64, userID int64) error {
	_, err := counterClient().ResetPrivateDialogCounter(ctx, &api.ResetPrivateDialogCounterRequest{
		DialogId: dialogID,
		OwnerId:  userID,
	})
	return err
}
