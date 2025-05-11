package service

import (
	"context"
	pb "github.com/Chamistery/subpub/api/pubsub"
	"github.com/Chamistery/subpub/pkg/subpub"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedPubSubServer
	bus subpub.SubPub
	log *zap.Logger
}

func NewServer(bus subpub.SubPub, log *zap.Logger) *Server {
	return &Server{bus: bus, log: log}
}

func (s *Server) Subscribe(req *pb.SubscribeRequest, stream pb.PubSub_SubscribeServer) error {
	key := req.GetKey()
	s.log.Info("new subscription", zap.String("key", key))

	sub, err := s.bus.Subscribe(key, func(msg interface{}) {
		// push to stream
		if err := stream.Send(&pb.Event{Data: msg.(string)}); err != nil {
			s.log.Error("send error", zap.Error(err))
		}
	})
	if err != nil {
		return grpcStatusError(codes.Internal, err)
	}
	// block until client cancels
	<-stream.Context().Done()
	sub.Unsubscribe()
	s.log.Info("subscription closed", zap.String("key", key))
	return nil
}

func (s *Server) Publish(ctx context.Context, req *pb.PublishRequest) (*empty.Empty, error) {
	key := req.GetKey()
	data := req.GetData()
	s.log.Info("publish event", zap.String("key", key), zap.String("data", data))
	if err := s.bus.Publish(key, data); err != nil {
		return nil, grpcStatusError(codes.Internal, err)
	}
	return &empty.Empty{}, nil
}

func grpcStatusError(code codes.Code, err error) error {
	return status.Errorf(code, "%v", err)
}
