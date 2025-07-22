package repository

import (
	"UrlShortner/model"

	"gorm.io/gorm"
)

type ShortUrlRepository struct {
	DB *gorm.DB
}

func (r *ShortUrlRepository) FindByShortCode(shortCode string) *model.ShortUrl {
	var shortUrl model.ShortUrl
	result := r.DB.Where("short_code = ?", shortCode).First(&shortUrl)
	if result.Error != nil {
		return nil
	}
	return &shortUrl
}

func (r *ShortUrlRepository) FindByLongUrl(longUrl string) *model.ShortUrl {
	var shortUrl model.ShortUrl
	result := r.DB.Where("long_url = ?", longUrl).First(&shortUrl)
	if result.Error != nil {
		return nil
	}
	return &shortUrl
}

func (r *ShortUrlRepository) Save(shortUrl *model.ShortUrl) {
	r.DB.Create(shortUrl)
}
