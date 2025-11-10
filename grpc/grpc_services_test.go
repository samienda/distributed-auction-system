package grpc_services

import (
	"context"
	"distributed-auction-system/grpc/gen"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuctionServer_CreateAuction(t *testing.T) {
	server := NewAuctionServer([]string{})
	ctx := context.Background()

	resp, err := server.CreateAuction(ctx, &gen.CreateAuctionRequest{
		AuctionId:     "auction1",
		OwnerId:       "owner1",
		Item:          "Laptop",
		StartingPrice: 100,
		ToPropagate:   false,
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Auction created successfully", resp.Message)

	resp, err = server.CreateAuction(ctx, &gen.CreateAuctionRequest{
		AuctionId:   "auction1",
		OwnerId:     "owner1",
		Item:        "Laptop",
		ToPropagate: false,
	})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Auction ID already exists", resp.Message)
}

func TestAuctionServer_PlaceBid(t *testing.T) {
	server := NewAuctionServer([]string{})
	ctx := context.Background()

	server.CreateAuction(ctx, &gen.CreateAuctionRequest{
		AuctionId:     "auction1",
		OwnerId:       "owner1",
		Item:          "Laptop",
		StartingPrice: 100,
		ToPropagate:   false,
	})

	resp, err := server.PlaceBid(ctx, &gen.BidRequest{
		AuctionId:   "auction1",
		BidderId:    "bidder1",
		Amount:      150,
		ToPropagate: false,
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Bid placed successfully", resp.Message)

	resp, err = server.PlaceBid(ctx, &gen.BidRequest{
		AuctionId:   "auction1",
		BidderId:    "bidder2",
		Amount:      120,
		ToPropagate: false,
	})

	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "Bid too low", resp.Message)
}

func TestAuctionServer_CloseAuction(t *testing.T) {
	server := NewAuctionServer([]string{})
	ctx := context.Background()

	server.CreateAuction(ctx, &gen.CreateAuctionRequest{
		AuctionId:     "auction1",
		OwnerId:       "owner1",
		Item:          "Laptop",
		StartingPrice: 100,
		ToPropagate:   false,
	})

	server.PlaceBid(ctx, &gen.BidRequest{
		AuctionId:   "auction1",
		BidderId:    "bidder1",
		Amount:      150,
		ToPropagate: false,
	})

	resp, err := server.CloseAuction(ctx, &gen.CloseAuctionRequest{
		AuctionId: "auction1",
		OwnerId:   "owner1",
	})

	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "Auction closed successfully", resp.Message)
	assert.Equal(t, "bidder1", resp.WinnerId)
	assert.Equal(t, float64(150), resp.FinalPrice)
}
