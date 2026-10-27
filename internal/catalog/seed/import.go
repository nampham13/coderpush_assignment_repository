package seed

import (
	"example.com/supplier-catalog/internal/catalog/repository/sqlite"
	"example.com/supplier-catalog/internal/catalog/service"
)

func Import(store *service.Store, source string) (int, error) {
	products, err := sqlite.ReadMockProducts(source)
	if err != nil {
		return 0, err
	}
	return store.SeedProducts(products)
}
