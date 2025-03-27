package dto

import "github.com/wassef911/astore/internal/delivery/models"

type ShopItem struct {
	ID          string  `json:"id" bson:"id,omitempty"`
	Title       string  `json:"title" bson:"title,omitempty"`
	Description string  `json:"description" bson:"description,omitempty"`
	Quantity    uint64  `json:"quantity" bson:"quantity,omitempty"`
	Price       float64 `json:"price" bson:"price,omitempty"`
}

type UpdateShoppingItemsReqDto struct {
	ShopItems []*models.ShopItem `json:"shopItems" bson:"shopItems,omitempty" validate:"required"`
}
