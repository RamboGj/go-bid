package product

import (
	"context"
	"go-bid/internal/validator"
	"time"
)

type CreateProductRequest struct {
	ProductName    string    `json:"product_name"`
	Description    string    `json:"description"`
	BasePriceCents int32     `json:"base_price_cents"`
	AuctionEnd     time.Time `json:"auction_end"`
}

const minAuctionDuration = 2 * time.Hour

func (req CreateProductRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator

	eval.CheckField(validator.NotBlank(req.ProductName), "product_name", "must not be blank")
	eval.CheckField(validator.NotBlank(req.Description), "description", "must not be blank")
	eval.CheckField(
		validator.MinChars(req.Description, 10) && validator.MaxChars(req.Description, 255),
		"description",
		"must contain between 10 and 255 characters",
	)
	eval.CheckField(req.BasePriceCents > 0, "base_price_cents", "must be greater than 0")
	eval.CheckField(time.Until(req.AuctionEnd) >= minAuctionDuration, "auction_end", "must be at least two hours duration")

	return eval
}
