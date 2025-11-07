package grpc_services

import (
	"context"
	"log"
	"sync"

	pb "distributed-auction-system/grpc/gen"

	"google.golang.org/grpc"
)

type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	auctions map[string]*Auction
	nodes    []string
	mu       sync.Mutex
}

func NewAuctionServer(nodes []string) *AuctionServer {
	return &AuctionServer{
		auctions: make(map[string]*Auction),
		nodes:    nodes,
	}
}

func (s *AuctionServer) CreateAuction(ctx context.Context, req *pb.CreateAuctionRequest) (*pb.CreateAuctionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.auctions[req.AuctionId]; exists {
		return &pb.CreateAuctionResponse{Success: false, Message: "Auction ID already exists"}, nil
	}

	s.auctions[req.AuctionId] = &Auction{
		AuctionID:    req.AuctionId,
		OwnerID:      req.OwnerId,
		Item:         req.Item,
		CurrentPrice: req.StartingPrice,
		IsActive:     true,
	}

	// propagate
	for _, node := range s.nodes {
		go func(nodeAddr string) {
			conn, err := grpc.Dial(nodeAddr, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to node %s: %v", nodeAddr, err)
				return
			}

			defer conn.Close()

			client := pb.NewAuctionServiceClient(conn)
			_, err = client.CreateAuction(ctx, req)
			if err != nil {
				log.Printf("Failed to propagate auction creation to %s: %v", nodeAddr, err)
			}
		}(node)
	}

	return &pb.CreateAuctionResponse{Success: true, Message: "Auction created successfully"}, nil
}

func (s *AuctionServer) PlaceBid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auction, exists := s.auctions[req.AuctionId]
	if !exists || !auction.IsActive {
		return &pb.BidResponse{Success: false, Message: "Auction not found or inactive"}, nil
	}

	auction.mu.Lock()
	defer auction.mu.Unlock()

	if req.Amount <= auction.CurrentPrice {
		return &pb.BidResponse{Success: false, Message: "Bid too low"}, nil
	}

	auction.CurrentPrice = req.Amount
	auction.HighestBidder = req.BidderId

	// propagate
	for _, node := range s.nodes {
		go func(nodeAddr string) {
			conn, err := grpc.Dial(nodeAddr, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to node %s: %v", nodeAddr, err)
				return
			}

			defer conn.Close()
			client := pb.NewAuctionServiceClient(conn)
			_, err = client.PlaceBid(ctx, req)
			if err != nil {
				log.Printf("Failed to propagate bid to %s: %v", nodeAddr, err)
			}
		}(node)
	}

	return &pb.BidResponse{Success: true, Message: "Bid placed successfully"}, nil
}

func (s *AuctionServer) CloseAuction(ctx context.Context, req *pb.CloseAuctionRequest) (*pb.CloseAuctionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	auction, exists := s.auctions[req.AuctionId]
	if !exists || !auction.IsActive {
		return &pb.CloseAuctionResponse{Success: false, Message: "Auction not found or inactive"}, nil
	}

	if auction.OwnerID != req.OwnerId {
		return &pb.CloseAuctionResponse{Success: false, Message: "Only the owner can close the auction"}, nil
	}

	auction.IsActive = false

	return &pb.CloseAuctionResponse{
		Success:    true,
		WinnerId:   auction.HighestBidder,
		FinalPrice: auction.CurrentPrice,
		Message:    "Auction closed successfully",
	}, nil
}