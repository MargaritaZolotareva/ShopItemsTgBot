package service

import (
	"VkScraper/model"
	"gorm.io/gorm"
	"math"
)

type BalloonService struct {
	DB *gorm.DB
}

func (s *BalloonService) GetAllProducts() ([]model.Balloon, error) {
	var products []model.Balloon
	if err := s.DB.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (s *BalloonService) CalculateNewPrice(balloon model.Balloon, heliumPrice int) int {
	return balloon.BalloonPrice + int(math.Ceil(float64(heliumPrice)*balloon.HeliumPortions)) + balloon.HiFloat*10
}
