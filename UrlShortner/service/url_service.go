package service

import (
	"UrlShortner/model"
	"UrlShortner/repository"

	"github.com/google/uuid"
)

type UrlService struct {
	Repo *repository.ShortUrlRepository
}

func (s *UrlService) CreateShortUrl(longUrl string) string {
	existing := s.Repo.FindByLongUrl(longUrl)
	if existing != nil {
		return existing.ShortCode
	}

	shortCode := uuid.New().String()[:6]
	shortUrl := &model.ShortUrl{
		LongUrl:   longUrl,
		ShortCode: shortCode,
	}
	s.Repo.Save(shortUrl)
	return shortCode
}

func (s *UrlService) GetOriginalUrl(shortCode string) *string {
	shortUrl := s.Repo.FindByShortCode(shortCode)
	if shortUrl == nil {
		return nil
	}
	return &shortUrl.LongUrl
}
