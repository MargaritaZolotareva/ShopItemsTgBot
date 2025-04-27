package handler

import (
	"VkScraper/model"
	"VkScraper/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func GenerateXLSXHandler() (string, []byte) {
	accessToken := os.Getenv("VK_ACCESS_TOKEN")
	groupID := os.Getenv("VK_GROUP_ID")
	albumIDs := strings.Split(os.Getenv("VK_ALBUM_IDS"), ",")

	var allProducts []model.Product

	for _, albumID := range albumIDs {
		products, err := GetProductsFromVK(accessToken, groupID, albumID)
		if err != nil {
			log.Fatalf("Ошибка загрузки товаров: %v", err)
		}
		allProducts = append(allProducts, products...)
	}

	fileName, fileBytes, err := service.WriteProductsToExcel(allProducts)
	if err != nil {
		log.Fatalf("Ошибка формирования Excel-файла: %v", err)
	}

	return fileName, fileBytes
}

func GetProductsFromVK(accessToken, groupID string, albumID string) ([]model.Product, error) {
	var products []model.Product
	url := fmt.Sprintf("https://api.vk.com/method/market.get?owner_id=-%s&album_id=%s&count=200&extended=1&v=5.131&access_token=%s", groupID, albumID, accessToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Response struct {
			Items []model.Item `json:"items"`
		} `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	for _, item := range result.Response.Items {
		price := ParsePrice(item.Price.Amount)
		variants := GetVariants(item, price)

		for _, variant := range variants {
			product := model.Product{
				ID:          variant.ID,
				Title:       item.Title,
				Price:       variant.Price,
				Description: item.Description,
			}
			products = append(products, product)
		}
	}

	return products, nil
}

func ParsePrice(priceStr string) float64 {
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	return price / 100 // Переводим в рубли
}

func GetVariants(item model.Item, basePrice float64) []model.Variant {
	var result []model.Variant
	var price float64

	for _, grid := range item.VariantsGrid {
		for _, variant := range grid.Variants {
			if variant.VariantID == 647 { // Без обработки
				price = basePrice
			} else if variant.VariantID == 648 { // С обработкой
				price = basePrice + 10
			}

			variantsData := model.Variant{
				ID:    variant.ItemID,
				Name:  variant.Name,
				Price: price,
			}
			result = append(result, variantsData)
		}
	}
	if len(result) == 0 {

		variantsData := model.Variant{
			ID:    item.ID,
			Name:  item.Title,
			Price: basePrice,
		}
		result = append(result, variantsData)
	}
	return result
}
