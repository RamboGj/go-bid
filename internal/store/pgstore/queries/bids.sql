-- name: CreateBid :one
INSERT INTO bids (
    product_id,
    bidder_id,
    bid_amount_cents
) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBidsByProductId :many
SELECT * FROM bids
WHERE product_id = $1
ORDER BY bid_amount_cents DESC
LIMIT $2 OFFSET $3;

-- name: GetHighestBidByProductId :one
SELECT * FROM bids
WHERE product_id = $1
ORDER BY bid_amount_cents DESC
LIMIT 1;
