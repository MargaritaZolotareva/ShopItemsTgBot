package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type VKMarketService struct {
	AccessToken string
	GroupID     string
}

func (s *VKMarketService) UpdateProductPrice(vkProductID int, newPrice int) error {
	apiURL := "https://api.vk.com/method/market.edit"

	data := url.Values{
		"owner_id":     {fmt.Sprintf("-%s", s.GroupID)},
		"item_id":      {fmt.Sprintf("%d", vkProductID)},
		"price":        {fmt.Sprintf("%d", newPrice)},
		"access_token": {s.AccessToken},
		"v":            {"5.131"},
	}

	urlWithParams := fmt.Sprintf("%s?%s", apiURL, data.Encode())

	resp, err := http.Get(urlWithParams)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["error"]; ok {
		log.Printf("Ошибка обновления товара %d: %v", vkProductID, result)
		return fmt.Errorf("Ошибка обновления товара %d", vkProductID)
	}

	log.Printf("Цена товара %d успешно обновлена до %d руб.", vkProductID, newPrice)
	return nil
}
