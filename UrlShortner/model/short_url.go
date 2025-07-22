package model

type ShortUrl struct {
	ID        uint   `gorm:"primaryKey"`
	LongUrl   string `gorm:"uniqueIndex"`
	ShortCode string `gorm:"uniqueIndex"`
}
