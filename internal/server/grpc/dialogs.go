package internalgrpc

import (
	"app/api"
	"app/service"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errInternalServerError = "internal server error"
)

type Dialogs struct {
	grpcServer *grpc.Server
	addr       string
	api.UnimplementedDialogsServer
}

func NewServer(host string, port int) *Dialogs {
	return &Dialogs{addr: net.JoinHostPort(host, strconv.Itoa(port))}
}

func (s *Dialogs) Start(_ context.Context) error {
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(loggingHandler))
	api.RegisterDialogsServer(s.grpcServer, s)

	lsn, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Err(err).Msgf("failed to listen grpc endpoint: %v", err)
		return err
	}

	log.Printf("starting grpc server on %s", s.addr)
	err = s.grpcServer.Serve(lsn)
	return err
}

func (s *Dialogs) Stop() {
	s.grpcServer.Stop()
}

func (s *Dialogs) Dialogs(_ context.Context, _ *emptypb.Empty) (*api.DialogsResponse, error) {
	dialogs, err := service.Dialogs()
	if err != nil {
		log.Err(err).Msgf("failed to get dialogs")
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	resp := api.DialogsResponse{}
	for _, dialog := range dialogs {
		resp.Dialogs = append(resp.Dialogs, &api.Dialog{
			Id:        dialog.ID,
			CreatorId: dialog.Creator.Id,
			Name:      dialog.Name,
		})
	}
	return &resp, nil
}

func (s *Dialogs) Dialog(_ context.Context, request *api.DialogRequest) (*api.DialogResponse, error) {
	dialog, err := service.Dialog(request.DialogId)
	if err != nil {
		log.Err(err).Msgf("failed to get answers of dialog '%d'", request.DialogId)
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &api.DialogResponse{Dialog: &api.Dialog{Id: dialog.ID, CreatorId: dialog.Creator.Id, Name: dialog.Name}}, nil
}

func (s *Dialogs) AddDialog(ctx context.Context, request *api.AddDialogRequest) (*emptypb.Empty, error) {
	err := service.AddDialog(request.CreatorId, request.Name)
	if err != nil {
		log.Err(err).Msgf("failed to get dialogs")
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &emptypb.Empty{}, nil
}

func (s *Dialogs) DialogAnswers(ctx context.Context, request *api.DialogAnswersRequest) (*api.DialogAnswersResponse, error) {
	answers, err := service.DialogAnswers(request.DialogId)
	if err != nil {
		log.Err(err).Msgf("failed to get answers of dialog '%d'", request.DialogId)
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	resp := api.DialogAnswersResponse{}
	for _, answer := range answers {
		resp.Answers = append(resp.Answers, &api.DialogAnswer{Id: answer.ID, Text: answer.Text})
	}
	return &resp, nil
}

func (s *Dialogs) AddDialogAnswer(ctx context.Context, request *api.AddDialogAnswerRequest) (*emptypb.Empty, error) {
	err := service.AddDialogAnswer(request.DialogId, request.CreatorId, request.Text)
	if err != nil {
		log.Err(err).Msgf("failed to add dialog answer")
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &emptypb.Empty{}, nil
}
