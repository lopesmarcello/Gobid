// Package product manages product use cases
package product

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lopesmarcello/gobid/internal/validator"
)

type CreateProductReq struct {
	SellerID    uuid.UUID `json:"seller_id"`
	ProductName string    `json:"product_name"`
	Description string    `json:"description"`
	Baseprice   float64   `json:"baseprice"`
	AuctionEnd  time.Time `json:"auction_end"`
}

const minAutctionDuration = 2 * time.Hour

func (req CreateProductReq) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "this field cannot be empty")
	eval.CheckField(validator.NotBlank(req.Description), "description", "this field cannot be empty")
	eval.CheckField(req.Baseprice > 0, "baseprice", "this field must be greater than 0")

	eval.CheckField(req.AuctionEnd.Sub(time.Now()) >= minAutctionDuration, "auction_end", "Auction should be longer than 2h")

	return eval
}
