package handler

import (
	"VkScraper/model"
	"fmt"
	"strings"
)

func GenerateYML(products []model.Product) (string, error) {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	sb.WriteString("\n<yml_catalog date=\"2025-03-13 02:32\">\n  <shop>\n")
	sb.WriteString("    <name>vk.com</name>\n    <company>vk.com</company>\n    <url>https://vk.com/</url>\n")
	sb.WriteString("    <currencies>\n      <currency id=\"RUB\" rate=\"1\"/>\n    </currencies>\n")
	sb.WriteString("    <offers>\n")

	for _, product := range products {
		sb.WriteString(fmt.Sprintf("      <offer id=\"%d\" available=\"true\">\n", product.ID))
		sb.WriteString(fmt.Sprintf("        <price>%.2f</price>\n", product.Price))
		sb.WriteString(fmt.Sprintf("        <currencyId>RUB</currencyId>\n"))
		sb.WriteString(fmt.Sprintf("        <categoryId>%d</categoryId>\n", ""))
		sb.WriteString(fmt.Sprintf("        <name>%s</name>\n", product.Title))

		sb.WriteString("      </offer>\n")
	}

	sb.WriteString("    </offers>\n  </shop>\n</yml_catalog>\n")

	return sb.String(), nil
}
