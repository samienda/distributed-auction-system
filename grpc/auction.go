package grpc_services

import "sync"

type Auction struct {
	AuctionID     string
	OwnerID       string
	Item          string
	CurrentPrice  float64
	HighestBidder string
	IsActive      bool
	mu            sync.Mutex
}