package file_formatter

import "ShopItemsTgBot/internal/model"

type FileFormatter interface {
	Generate(products []model.Product) (string, []byte, error)
}
