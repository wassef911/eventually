package models

import (
	"fmt"
)

type ShopItem struct {
	ID          string  `json:"id" bson:"id,omitempty"`
	Title       string  `json:"title" bson:"title,omitempty"`
	Description string  `json:"description" bson:"description,omitempty"`
	Quantity    uint64  `json:"quantity" bson:"quantity,omitempty"`
	Price       float64 `json:"price" bson:"price,omitempty"`
}

func (s *ShopItem) String() string {
	return fmt.Sprintf("ID: {%s}, Title: {%s}, Description: {%s}, Quantity: {%v}, Price: {%v},",
		s.ID,
		s.Title,
		s.Description,
		s.Quantity,
		s.Price,
	)
}
