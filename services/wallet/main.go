// The wallet service holds balances and moves money.
// It serves gRPC on GRPC_PORT (default 50051) and ops endpoints on OPS_PORT (default 9090).
package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	walletv1 "github.com/Utkarsh-262003/wallet-ledger/gen/wallet/v1"
	"github.com/Utkarsh-262003/wallet-ledger/internal/platform"
)

func main() {
	log := platform.NewLogger("wallet")
	dsn := platform.MustEnv("DATABASE_URL")
	grpcPort := platform.EnvOr("GRPC_PORT", "50051")

	// pgxpool.New checks the URL but does not connect yet.
	// The readiness check below is what proves the database is reachable.
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Error("invalid DATABASE_URL", "err", err)
		os.Exit(1)
	}

	ops := platform.NewOps()
	ops.AddCheck("postgres", pool.Ping)
	ops.Start(log)

	store := &Store{pool: pool}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(metricsInterceptor))
	walletv1.RegisterWalletServiceServer(grpcServer, &walletServer{store: store, log: log})

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Error("cannot listen", "port", grpcPort, "err", err)
		os.Exit(1)
	}
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server stopped", "err", err)
			os.Exit(1)
		}
	}()
	log.Info("wallet gRPC server listening", "port", grpcPort)

	sig := platform.WaitForSignal()
	platform.BeginShutdown(ops, log, sig)

	// Stop in order: stop taking RPCs (finishing the ones in flight), then close connections.
	grpcServer.GracefulStop()
	pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = ops.Shutdown(ctx)
	log.Info("shutdown complete")
}