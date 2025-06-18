package file_formatter

import "errors"

func NewFileFormatter(formatType string) (FileFormatter, error) {
	switch formatType {
	case "yandex":
		return &YandexFileFormatter{}, nil
	case "2gis":
		return &TwoGisFileFormatter{}, nil
	default:
		return nil, errors.New("unknown format type")
	}
}
