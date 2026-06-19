package services

import (
	"context"
	"errors"
	"go-bid/internal/store/pgstore"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBidIsTooLow = errors.New("bid amount is too low")
)

type BidsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewBidsService(pool *pgxpool.Pool) BidsService {
	return BidsService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (bs *BidsService) PlaceBid(ctx context.Context, product_id, bidder_id uuid.UUID, amount_cents int32) (pgstore.Bid, error) {
	// Amount must be greater than greatest bid amount for this given product id
	// Amount > Base price

	product, err := bs.queries.GetProductById(ctx, product_id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	highestBid, err := bs.queries.GetHighestBidByProductId(ctx, product_id)

	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.BasePriceCents >= amount_cents || highestBid.BidAmountCents >= amount_cents {
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = bs.queries.CreateBid(ctx, pgstore.CreateBidParams{
		ProductID:      product_id,
		BidderID:       bidder_id,
		BidAmountCents: amount_cents,
	})

	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
