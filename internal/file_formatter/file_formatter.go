package file_formatter

import "VkScraper/model"

type FileFormatter interface {
	Generate(products []model.Product) (string, []byte, error)
}
