package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"

	pb "github.com/Chamistery/subpub/api/pubsub"
	"github.com/Chamistery/subpub/pkg/subpub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeBus struct {
	subscribeErr error
	publishErr   error
	callbacks    map[string]subpub.MessageHandler
	published    []struct{ key, data string }
}

func newFakeBus() *fakeBus {
	return &fakeBus{callbacks: make(map[string]subpub.MessageHandler)}
}

func (f *fakeBus) Subscribe(subject string, cb subpub.MessageHandler) (subpub.Subscription, error) {
	if f.subscribeErr != nil {
		return nil, f.subscribeErr
	}
	f.callbacks[subject] = cb
	return &fakeSub{}, nil
}

func (f *fakeBus) Publish(subject string, msg interface{}) error {
	if f.publishErr != nil {
		return f.publishErr
	}
	f.published = append(f.published, struct{ key, data string }{subject, msg.(string)})
	return nil
}

func (f *fakeBus) Close(ctx context.Context) error { return nil }

type fakeSub struct{}

func (s *fakeSub) Unsubscribe() {}

type fakeStream struct {
	ctx context.Context
	grpc.ServerStream
	sent []*pb.Event
	err  error
}

func (fs *fakeStream) Send(ev *pb.Event) error {
	if fs.err != nil {
		return fs.err
	}
	fs.sent = append(fs.sent, ev)
	return nil
}

func (fs *fakeStream) Context() context.Context       { return fs.ctx }
func (f *fakeStream) SetHeader(md metadata.MD) error  { return nil }
func (f *fakeStream) SendHeader(md metadata.MD) error { return nil }
func (f *fakeStream) SetTrailer(md metadata.MD)       {}

func TestPublishSuccess(t *testing.T) {
	bus := newFakeBus()
	srv := NewServer(bus, zap.NewNop())

	resp, err := srv.Publish(context.Background(), &pb.PublishRequest{Key: "k", Data: "d"})
	require.NoError(t, err)
	require.IsType(t, &emptypb.Empty{}, resp)
	require.Len(t, bus.published, 1)
	require.Equal(t, "k", bus.published[0].key)
	require.Equal(t, "d", bus.published[0].data)
}

func TestPublishError(t *testing.T) {
	bus := newFakeBus()
	bus.publishErr = errors.New("fail")
	server := NewServer(bus, zap.NewNop())

	_, err := server.Publish(context.Background(), &pb.PublishRequest{Key: "k", Data: "d"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
}

func TestSubscribeSuccess(t *testing.T) {
	bus := newFakeBus()
	server := NewServer(bus, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	stream := &fakeStream{ctx: ctx}

	go func() {
		// give Subscribe time to register
		time.Sleep(10 * time.Millisecond)
		bus.callbacks["key"]("m1")
		bus.callbacks["key"]("m2")
		cancel()
	}()

	err := server.Subscribe(&pb.SubscribeRequest{Key: "key"}, stream)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(stream.sent) != 2 {
		t.Errorf("expected 2 events, got %d", len(stream.sent))
	}
	if stream.sent[0].Data != "m1" || stream.sent[1].Data != "m2" {
		t.Errorf("unexpected events: %v", stream.sent)
	}
}

func TestSubscribeError(t *testing.T) {
	bus := newFakeBus()
	bus.subscribeErr = errors.New("fail sub")
	server := NewServer(bus, zap.NewNop())
	stream := &fakeStream{ctx: context.Background()}

	err := server.Subscribe(&pb.SubscribeRequest{Key: "key"}, stream)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Errorf("expected code %v, got %v", codes.Internal, st.Code())
	}
}
