package models

import (
	"fmt"
	"time"
)

type OrderProjection struct {
	ID              string      `json:"id" bson:"_id,omitempty"`
	OrderID         string      `json:"orderId,omitempty" bson:"orderId,omitempty"`
	ShopItems       []*ShopItem `json:"shopItems,omitempty" bson:"shopItems,omitempty"`
	AccountEmail    string      `json:"accountEmail,omitempty" bson:"accountEmail,omitempty" validate:"required,email"`
	DeliveryAddress string      `json:"deliveryAddress,omitempty" bson:"deliveryAddress,omitempty"`
	CancelReason    string      `json:"cancelReason,omitempty" bson:"cancelReason,omitempty"`
	TotalPrice      float64     `json:"totalPrice,omitempty" bson:"totalPrice,omitempty"`
	DeliveredTime   time.Time   `json:"deliveredTime,omitempty" bson:"deliveredTime,omitempty"`
	Paid            bool        `json:"paid,omitempty" bson:"paid,omitempty"`
	Submitted       bool        `json:"submitted,omitempty" bson:"submitted,omitempty"`
	Completed       bool        `json:"completed,omitempty" bson:"completed,omitempty"`
	Canceled        bool        `json:"canceled,omitempty" bson:"canceled,omitempty"`
	Payment         Payment     `json:"payment,omitempty" bson:"payment,omitempty"`
}

func (o *OrderProjection) String() string {
	return fmt.Sprintf("ID: {%s}, ShopItems: {%+v}, Paid: {%v}, Submitted: {%v}, "+
		"Completed: {%v}, Canceled: {%v}, CancelReason: {%s}, TotalPrice: {%v}, AccountEmail: {%s}, DeliveryAddress: {%s}, DeliveredTime: {%s}, Payment: {%s}",
		o.ID,
		o.ShopItems,
		o.Paid,
		o.Submitted,
		o.Completed,
		o.Canceled,
		o.CancelReason,
		o.TotalPrice,
		o.AccountEmail,
		o.DeliveryAddress,
		o.DeliveredTime.UTC().String(),
		o.Payment.String(),
	)
}
