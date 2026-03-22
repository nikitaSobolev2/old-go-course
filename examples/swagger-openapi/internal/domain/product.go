package domain // сущность товара каталога

import "errors"

var ErrNotFound = errors.New("not found") // товар с данным id отсутствует

// Product is a catalog item (bounded context: catalog).
type Product struct {
	ID    string
	Name  string
	Price float64
}
