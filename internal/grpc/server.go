package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		msg := "cannot convert user id to string"
		return nil, status.Error(codes.Internal, msg)
	}

	err := s.service.StoreData(ctx, id, string(in.GetUrl()), userId)

	if err != nil && !s.service.IsErrAsDuplicate(err) {
		s.logger.Errorf("store error: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Если дубликат, то тогда извлекаем по значению
	if err != nil && s.service.IsErrAsDuplicate(err) {
		s.logger.Debug("Duplicate, try getting exist reccord")

		id, err = s.service.GetShortByOrigin(ctx, in.GetUrl())

		if err != nil {
			s.logger.Errorf("store error: %s", err)

			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	full := s.cfg.BaseUrl + "/" + id

	var resp pb.URLShortenResponse

	resp.SetResult(full)

	s.service.AuditNotify(audit.ShortenAction, userId, string(in.GetUrl()))

	return &resp, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.CtxTimeout)

	defer cancel()

	url, err := s.service.GetUrl(ctx, in.GetId())
	var deletedError service.DeletedUrlError

	if errors.As(err, &deletedError) {
		err := errors.New("url was deleted")
		s.logger.Error(err)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if err != nil {
		s.logger.Errorf("get data error: %s", err)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	if url == "" {
		err := errors.New("url not found")
		s.logger.Error(err)
		return nil, status.Error(codes.NotFound, err.Error())
	}

	userId := s.service.GetUserIdFromCtx(ctx)

	s.service.AuditNotify(audit.FollowAction, userId, url)

	var resp pb.URLExpandResponse

	resp.SetResult(url)

	return &resp, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, in *emptypb.Empty) (*pb.UserURLsResponse, error) {
	return nil, nil
}
