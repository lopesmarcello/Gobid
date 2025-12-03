// Package services bids services
package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lopesmarcello/gobid/internal/store/pgstore"
)

type BidsService struct {
	pool    *pgxpool.Pool
	queries *pgstore.Queries
}

func NewBidsService(pool *pgxpool.Pool) BidsService {
	return BidsService{
		pool,
		pgstore.New(pool),
	}
}

var ErrBidIsTooLow = errors.New("the bid value is too low")

func (bs *BidsService) PlaceBid(ctx context.Context, productID, bidderID uuid.UUID, amount float64) (pgstore.Bid, error) {
	// amount > previous amount
	// amount > baseprice
	//
	product, err := bs.queries.GetProductById(ctx, productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	highestBid, err := bs.queries.GetHighestBidByProductId(ctx, productID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Bid{}, err
		}
	}

	if product.Baseprice >= amount || highestBid.BidAmount >= amount {
		return pgstore.Bid{}, ErrBidIsTooLow
	}

	highestBid, err = bs.queries.CreateBid(ctx,
		pgstore.CreateBidParams{
			ProductID: productID,
			BidderID:  bidderID,
			BidAmount: amount,
		})
	if err != nil {
		return pgstore.Bid{}, err
	}

	return highestBid, nil
}
