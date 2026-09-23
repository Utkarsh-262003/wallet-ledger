package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	walletv1 "github.com/Utkarsh-262003/wallet-ledger/gen/wallet/v1"
)

// walletServer implements the gRPC service defined in proto/wallet/v1/wallet.proto.
// Embedding UnimplementedWalletServiceServer means every method we have not
// written yet answers with "Unimplemented" instead of failing to compile.
type walletServer struct {
	walletv1.UnimplementedWalletServiceServer
	store *Store
	log   *slog.Logger
}

func (s *walletServer) CreateWallet(ctx context.Context, req *walletv1.CreateWalletRequest) (*walletv1.CreateWalletResponse, error) {
	if err := requireUUID(req.GetUserId(), "user_id"); err != nil {
		return nil, err
	}
	w, err := s.store.CreateWallet(ctx, req.GetUserId())
	if err != nil {
		return nil, s.internal("create wallet", err)
	}
	return &walletv1.CreateWalletResponse{Wallet: w.toProto()}, nil
}

func (s *walletServer) GetWallet(ctx context.Context, req *walletv1.GetWalletRequest) (*walletv1.GetWalletResponse, error) {
	if err := requireUUID(req.GetUserId(), "user_id"); err != nil {
		return nil, err
	}
	w, err := s.store.GetWalletByUser(ctx, req.GetUserId())
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}
	if err != nil {
		return nil, s.internal("get wallet", err)
	}
	return &walletv1.GetWalletResponse{Wallet: w.toProto()}, nil
}

// internal logs the real error and returns a generic one.
// Database details stay in our logs and never reach the caller.
func (s *walletServer) internal(op string, err error) error {
	s.log.Error("internal error", "op", op, "err", err)
	return status.Error(codes.Internal, "internal error")
}

func requireUUID(value, field string) error {
	if _, err := uuid.Parse(value); err != nil {
		return status.Errorf(codes.InvalidArgument, "%s must be a UUID", field)
	}
	return nil
}

func (w Wallet) toProto() *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:           w.ID,
		UserId:       w.UserID,
		BalanceCents: w.BalanceCents,
		Currency:     w.Currency,
		Version:      w.Version,
		CreatedAt:    w.CreatedAt.UTC().Format(time.RFC3339),
	}
}