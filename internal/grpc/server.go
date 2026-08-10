package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/spider4216/tinyurl/internal/audit"
	"github.com/spider4216/tinyurl/internal/config"
	"github.com/spider4216/tinyurl/internal/service"
	pb "github.com/spider4216/tinyurl/proto"
)

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	cfg     *config.Config
	service service.Service
	logger  *zap.SugaredLogger
}

func New(
	cfg *config.Config,
	service service.Service,
	logger *zap.SugaredLogger,
) *ShortenerServer {
	return &ShortenerServer{
		cfg:     cfg,
		service: service,
		logger:  logger,
	}

}

func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	s.logger.Info("Catch grp on ShortenURL")

	ctx, cancel := context.WithTimeout(ctx, s.cfg.CtxTimeout)

	defer cancel()

	id := s.service.GenerateId()

	userId := s.service.GetUserIdFromCtx(ctx)

	if userId == "" {
		err := errors.New("cannot convert user id to string")
		s.logger.Error(err)
		return nil, err
	}

	err := s.service.StoreData(ctx, id, string(in.GetUrl()), userId)

	if err != nil && !s.service.IsErrAsDuplicate(err) {
		s.logger.Errorf("store error: %s", zap.Error(err))
		return nil, err
	}

	// Если дубликат, то тогда извлекаем по значению
	if err != nil && s.service.IsErrAsDuplicate(err) {
		s.logger.Debug("Duplicate, try getting exist reccord")

		id, err = s.service.GetShortByOrigin(ctx, in.GetUrl())

		if err != nil {
			s.logger.Errorf("store error: %s", zap.Error(err))

			return nil, err
		}
	}

	full := s.cfg.BaseUrl + "/" + id

	var resp pb.URLShortenResponse

	resp.SetResult(full)

	s.service.AuditNotify(audit.ShortenAction, userId, string(in.GetUrl()))

	return &resp, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	return nil, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, in *emptypb.Empty) (*pb.UserURLsResponse, error) {
	return nil, nil
}
