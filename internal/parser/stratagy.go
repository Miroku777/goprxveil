package parser

import (
	"github.com/Miroku77/goprxveil/internal/models"
	"github.com/go-rod/rod"
)

type Parser interface {
	ParseAll(urlstr string) (models.Product, int64, error)
	ParsePrice(page *rod.Page, urlstr string) (int64, error)
}
