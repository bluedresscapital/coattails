package poncho

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

type OrderAPI interface {
	GetOrders() ([]wardrobe.Order, error)
}

func ReloadOrders(api OrderAPI) error {
	orders, err := api.GetOrders()
	if err != nil {
		return err
	}
	for _, o := range orders {
		log.Printf("%v", o)
	}
	return nil
}
