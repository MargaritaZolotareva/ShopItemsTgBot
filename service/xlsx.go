package service

import (
	"VkScraper/model"
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

func WriteProductsToExcel(products []model.Product) (string, []byte, error) {
	f := excelize.NewFile()

	sheet := "Products"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "ID")
	f.SetCellValue(sheet, "B1", "Title")
	f.SetCellValue(sheet, "C1", "Price")

	for i, product := range products {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), product.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), product.Title)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), product.Price)
		row++
	}

	fileName := "products.xlsx"
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
