package internalgrpc

import (
	"app/api"
	"app/service"
	"context"
	"github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
)

const (
	errInternalServerError = "internal server error"
)

type Dialogs struct {
	grpcServer *grpc.Server
	addr       string
	api.UnimplementedDialogsServer
	zipkinUrl string
	tracer    *zipkin.Tracer
}

func NewServer(host string, port int, zipkinUrl string) *Dialogs {
	return &Dialogs{addr: net.JoinHostPort(host, strconv.Itoa(port)), zipkinUrl: zipkinUrl}
}

func (s *Dialogs) Start(_ context.Context) error {
	reporter := http.NewReporter(s.zipkinUrl)
	s.tracer, _ = zipkin.NewTracer(reporter)
	s.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(loggingHandler), grpc.StatsHandler(zipkingrpc.NewServerHandler(s.tracer)))
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

func (s *Dialogs) Dialogs(ctx context.Context, _ *emptypb.Empty) (*api.DialogsResponse, error) {
	dialogs, err := service.Dialogs(ctx)
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

func (s *Dialogs) Dialog(ctx context.Context, request *api.DialogRequest) (*api.DialogResponse, error) {
	span, _ := s.tracer.StartSpanFromContext(ctx, "mysql")
	dialog, err := service.Dialog(ctx, request.DialogId)
	span.Finish()
	if err != nil {
		log.Err(err).Msgf("failed to get answers of dialog '%d'", request.DialogId)
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &api.DialogResponse{Dialog: &api.Dialog{Id: dialog.ID, CreatorId: dialog.Creator.Id, Name: dialog.Name}}, nil
}

func (s *Dialogs) AddDialog(ctx context.Context, request *api.AddDialogRequest) (*emptypb.Empty, error) {
	span, _ := s.tracer.StartSpanFromContext(ctx, "mysql")
	err := service.AddDialog(ctx, request.CreatorId, request.Name)
	span.Finish()
	if err != nil {
		log.Err(err).Msgf("failed to get dialogs")
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &emptypb.Empty{}, nil
}

func (s *Dialogs) DialogAnswers(ctx context.Context, request *api.DialogAnswersRequest) (*api.DialogAnswersResponse, error) {
	span, _ := s.tracer.StartSpanFromContext(ctx, "mysql")
	answers, err := service.DialogAnswers(ctx, request.DialogId)
	span.Finish()
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
	span, _ := s.tracer.StartSpanFromContext(ctx, "mysql")
	err := service.AddDialogAnswer(ctx, request.DialogId, request.CreatorId, request.Text)
	span.Finish()
	if err != nil {
		log.Err(err).Msgf("failed to add dialog answer")
		return nil, status.Errorf(codes.Internal, errInternalServerError)
	}
	return &emptypb.Empty{}, nil
}
