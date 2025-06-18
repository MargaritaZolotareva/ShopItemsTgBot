package service

import (
	"VkScraper/model"
	"bytes"
	"fmt"
	"github.com/xuri/excelize/v2"
	"time"
)

func WriteProductsToExcel(products []model.Product) (string, []byte, error) {
	f := excelize.NewFile()

	sheet := "Products"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "product_id")
	f.SetCellValue(sheet, "B1", "name(ru-ru)")
	f.SetCellValue(sheet, "C1", "categories")
	f.SetCellValue(sheet, "D1", "sku")
	f.SetCellValue(sheet, "E1", "photo")
	f.SetCellValue(sheet, "F1", "ean")
	f.SetCellValue(sheet, "G1", "jan")
	f.SetCellValue(sheet, "H1", "isbn")
	f.SetCellValue(sheet, "I1", "mpn")
	f.SetCellValue(sheet, "J1", "location")
	f.SetCellValue(sheet, "K1", "quantity")
	f.SetCellValue(sheet, "L1", "model")
	f.SetCellValue(sheet, "M1", "manufacturer")
	f.SetCellValue(sheet, "N1", "image_name")
	f.SetCellValue(sheet, "O1", "shipping")
	f.SetCellValue(sheet, "P1", "price")
	f.SetCellValue(sheet, "Q1", "points")
	f.SetCellValue(sheet, "R1", "date_added")
	f.SetCellValue(sheet, "S1", "date_modified")
	f.SetCellValue(sheet, "T1", "date_available")
	f.SetCellValue(sheet, "U1", "weight")
	f.SetCellValue(sheet, "V1", "weight_unit")
	f.SetCellValue(sheet, "W1", "length")
	f.SetCellValue(sheet, "X1", "width")
	f.SetCellValue(sheet, "Y1", "height")
	f.SetCellValue(sheet, "Z1", "length_unit")
	f.SetCellValue(sheet, "AA1", "status")
	f.SetCellValue(sheet, "AB1", "tax_class_id")
	f.SetCellValue(sheet, "AC1", "description(ru-ru)")
	f.SetCellValue(sheet, "AD1", "meta_title(ru-ru)")
	f.SetCellValue(sheet, "AE1", "meta_description(ru-ru)")
	f.SetCellValue(sheet, "AF1", "meta_keywords(ru-ru)")
	f.SetCellValue(sheet, "AG1", "stock_status_id")
	f.SetCellValue(sheet, "AH1", "store_ids")
	f.SetCellValue(sheet, "AI1", "layout")
	f.SetCellValue(sheet, "AJ1", "related_ids")
	f.SetCellValue(sheet, "AK1", "tags(ru-ru)")
	f.SetCellValue(sheet, "AL1", "sort_order")
	f.SetCellValue(sheet, "AM1", "subtract")
	f.SetCellValue(sheet, "AN1", "minimum")
	f.SetCellValue(sheet, "AO1", "upc")

	currentTime := time.Now()
	for i, product := range products {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), product.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), product.Title)
		f.SetCellValue(sheet, fmt.Sprintf("AC%d", row), product.Description)
		f.SetCellValue(sheet, fmt.Sprintf("P%d", row), product.Price)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), product.LargestPhoto)
		f.SetCellValue(sheet, fmt.Sprintf("K%d", row), 100)
		f.SetCellValue(sheet, fmt.Sprintf("R%d", row), currentTime.Format("2006-01-02 15:04:05"))
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
