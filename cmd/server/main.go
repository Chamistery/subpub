package main

import (
	"context"
	"fmt"
	"github.com/Chamistery/subpub/internal/logger"
	"go.uber.org/zap/zapcore"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/Chamistery/subpub/api/pubsub"
	"github.com/Chamistery/subpub/cmd/server/config"
	svc "github.com/Chamistery/subpub/internal/service"
	"github.com/Chamistery/subpub/pkg/subpub"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %v", err)
		os.Exit(1)
	}

	lvl := zapcore.DebugLevel
	logger.Init(lvl)
	defer logger.Log.Sync()
	logger.Log.Info("config loaded", zap.Any("cfg", cfg))

	bus := subpub.NewSubPub()
	defer bus.Close(context.Background())

	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}
	grpcServer := grpc.NewServer()
	handler := svc.NewServer(bus, logger.Log)
	pb.RegisterPubSubServer(grpcServer, handler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Log.Info("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		go grpcServer.GracefulStop()
		bus.Close(ctx)
	}()

	logger.Log.Info("starting gRPC server", zap.String("addr", cfg.Port))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("serve error", zap.Error(err))
	}
}
