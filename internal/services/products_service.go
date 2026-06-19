package services

import (
	"context"
	"go-bid/internal/store/pgstore"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewProductService(pool *pgxpool.Pool) ProductService {
	return ProductService{
		pool:    pool,
		queries: pgstore.New(pool),
	}
}

func (ps *ProductService) CreateProduct(
	ctx context.Context,
	sellerId uuid.UUID,
	productName string,
	description string,
	basePriceCents int32,
	auctionEnd time.Time,
) (uuid.UUID, error) {
	id, err := ps.queries.CreateProduct(ctx, pgstore.CreateProductParams{
		SellerID:       sellerId,
		ProductName:    productName,
		Description:    description,
		BasePriceCents: basePriceCents,
		AuctionEnd:     auctionEnd,
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
