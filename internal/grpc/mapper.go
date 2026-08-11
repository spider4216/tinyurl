package grpc

import (
	"github.com/spider4216/tinyurl/internal/models"
	pb "github.com/spider4216/tinyurl/proto"
)

func (s *ShortenerServer) mapListUrlsResp(urls []models.UrlItem) *pb.UserURLsResponse {
	var resp pb.UserURLsResponse
	var urlData []*pb.URLData

	for _, url := range urls {
		item := &pb.URLData{}

		item.SetOriginalUrl(url.OriginarUrl)
		item.SetShortUrl(url.ShortUrl)

		urlData = append(urlData, item)
	}

	resp.SetUrl(urlData)

	return &resp
}
