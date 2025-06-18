package file_formatter

import (
	"VkScraper/model"
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
)

type TwoGisFileFormatter struct{}

func (gisf *TwoGisFileFormatter) Generate(products []model.Product) (string, []byte, error) {
	f := excelize.NewFile()

	sheet := "Лист 1"
	f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")
	f.SetCellValue(sheet, "A1", "Наименование товара")
	f.SetCellValue(sheet, "B1", "Цена")
	f.SetCellValue(sheet, "C1", "Цена от")
	f.SetCellValue(sheet, "D1", "Цена до")
	f.SetCellValue(sheet, "E1", "Категория")
	f.SetCellValue(sheet, "F1", "Ссылка на товар на сайте магазина")
	f.SetCellValue(sheet, "G1", "Ссылка на картинку")
	f.SetCellValue(sheet, "H1", "Описание")

	for i, product := range products {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), product.Title)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), product.Price)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), product.Category)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), product.LargestPhoto)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), product.Description)
		row++
	}

	fileName := "products_2gis.xlsx"
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
