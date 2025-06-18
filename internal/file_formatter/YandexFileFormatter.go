package file_formatter

import (
	"VkScraper/model"
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
)

type YandexFileFormatter struct{}

func (yf *YandexFileFormatter) Generate(products []model.Product) (string, []byte, error) {
	f := excelize.NewFile()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "Категория")
	f.SetCellValue(sheet, "B1", "Название")
	f.SetCellValue(sheet, "C1", "Идентификатор")
	f.SetCellValue(sheet, "D1", "Описание")
	f.SetCellValue(sheet, "E1", "Короткое описание")
	f.SetCellValue(sheet, "F1", "Цена")
	f.SetCellValue(sheet, "G1", "Фото")

	for i, product := range products {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), product.Category)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), product.Title)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), product.ID)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), product.Description)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), product.Description)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), product.Price)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), product.LargestPhoto)
		row++
	}

	fileName := "products_yandex.xlsx"
	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return "", nil, err
	}
	if err != nil {
		return "", nil, fmt.Errorf("не удалось сохранить файл: %v", err)
	}

	return fileName, buf.Bytes(), nil
}
