//go:build !dialogsservice
// +build !dialogsservice

package service

import (
	"app/api"
	"app/models"
	"context"
	"errors"
	"github.com/beego/beego/v2/client/orm"
	"github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
)

var (
	grpcDialogsConn   *grpc.ClientConn
	grpcDialogsClient api.DialogsClient
)

func SetupGrpcDialogs(host string, port int, tracer *zipkin.Tracer) error {
	var err error
	grpcDialogsConn, err = grpc.Dial(
		net.JoinHostPort(host, strconv.Itoa(port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(zipkingrpc.NewClientHandler(tracer)),
	)
	if err != nil {
		return err
	}

	grpcDialogsClient = api.NewDialogsClient(grpcDialogsConn)
	return nil
}

func StopGrpcDialogs() {
	if grpcDialogsConn != nil {
		grpcDialogsConn.Close()
	}
}

func dialogsClient() api.DialogsClient {
	return grpcDialogsClient
}

func Dialogs(ctx context.Context) ([]models.Dialog, error) {
	resp, err := dialogsClient().Dialogs(ctx, &emptypb.Empty{})
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	dialogs := make([]models.Dialog, 0, len(resp.GetDialogs()))
	for _, d := range resp.GetDialogs() {
		dialogs = append(dialogs, models.Dialog{ID: d.Id, Name: d.Name})
	}
	return dialogs, nil
}

func Dialog(ctx context.Context, id int64) (models.Dialog, error) {
	resp, err := dialogsClient().Dialog(ctx, &api.DialogRequest{DialogId: id})
	if err != nil {
		return models.Dialog{}, err
	}
	return models.Dialog{
		ID:      resp.GetDialog().GetId(),
		Name:    resp.GetDialog().GetName(),
		Creator: models.User{Id: resp.GetDialog().GetCreatorId()},
	}, nil
}

func AddDialog(ctx context.Context, creatorID int64, name string) error {
	_, err := dialogsClient().AddDialog(ctx, &api.AddDialogRequest{CreatorId: creatorID, Name: name})
	return err
}

func DialogAnswers(ctx context.Context, dialogID int64) ([]models.DialogAnswer, error) {
	resp, err := dialogsClient().DialogAnswers(ctx, &api.DialogAnswersRequest{DialogId: dialogID})
	if err != nil {
		return nil, err
	}

	answers := make([]models.DialogAnswer, len(resp.GetAnswers()))
	for i, a := range resp.GetAnswers() {
		answers[i] = models.DialogAnswer{ID: a.Id, Text: a.Text}
	}
	return answers, nil
}

func AddDialogAnswer(ctx context.Context, dialogID int64, creatorID int64, text string) error {
	_, err := dialogsClient().AddDialogAnswer(
		ctx,
		&api.AddDialogAnswerRequest{CreatorId: creatorID, DialogId: dialogID, Text: text},
	)
	return err
}
