package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/grpc/interceptors"
	"github.com/spider4216/tinyurl/internal/logger"
	"github.com/spider4216/tinyurl/internal/repository"
	"github.com/spider4216/tinyurl/internal/service"
	"github.com/spider4216/tinyurl/internal/storage"
	pb "github.com/spider4216/tinyurl/proto"
)

func TestShortenURL(t *testing.T) {
	cfg, err := config.New()
	cfg.SignKey = "abc"

	require.NoError(t, err)
	logger, err := logger.InitZap("debug")
	require.NoError(t, err)

	store, err := storage.New(storage.MapDriver, cfg, logger)
	require.NoError(t, err)

	r := repository.New(store)

	event := audit.NewAuditEvent(logger)
	s := service.New(r, logger, event)

	interceptors := interceptors.New(cfg, s, logger)
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(interceptors.WithAuth))

	reflection.Register(grpcSrv)
	myGrpcSrv := New(cfg, s, logger)

	listen, err := net.Listen("tcp", ":1212")

	require.NoError(t, err)

	pb.RegisterShortenerServiceServer(grpcSrv, myGrpcSrv)

	go func() {
		if err := grpcSrv.Serve(listen); err != nil {
			t.Errorf("server error: %v", err)
		}
	}()

	defer grpcSrv.Stop()

	ctx := context.Background()

	// Устанавливаем соединение с сервером
	conn, err := grpc.NewClient(listen.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	require.NoError(t, err)
	defer func() {
		err := conn.Close()
		require.NoError(t, err)
	}()

	// Формируем ключ
	userID := uuid.NewString()
	sign, err := s.SignVal(userID, cfg.SignKey)
	require.NoError(t, err)
	// Готовим метаданные
	md := metadata.New(map[string]string{"authorization": userID + "." + sign})
	ctx = metadata.NewOutgoingContext(ctx, md)

	c := pb.NewShortenerServiceClient(conn)

	var req pb.URLShortenRequest

	req.SetUrl("http://example.com")

	resp, err := c.ShortenURL(ctx, &req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.GetResult())
}
