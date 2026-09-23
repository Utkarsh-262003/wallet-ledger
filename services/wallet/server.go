package main

import (
	"log/slog"

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