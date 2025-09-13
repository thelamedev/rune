package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/thelamedev/rune/internal/crypto"
	"github.com/thelamedev/rune/internal/seal"
	"github.com/thelamedev/rune/internal/server"
	"github.com/thelamedev/rune/internal/storage"
)

func main() {
	log.Println("--- Starting Rune Server ---")

	dbPath := "rune.db"
	store, err := storage.NewBoltStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Fatalf("Failed to close storage: %v", err)
		}
	}()

	keyShares, keyThreshold := 5, 3
	sealManager := seal.New(keyShares, keyThreshold)

	log.Println("Initializing and unsealing the vault")
	shares, err := sealManager.GenerateKeys(context.Background())
	if err != nil {
		log.Fatalf("Failed to generate seal keys: %v", err)
	}

	for i := range keyThreshold {
		if _, _, err := sealManager.Unseal(context.Background(), shares[i]); err != nil {
			log.Fatalf("Failed to unseal vault: %v", err)
		}
	}

	if !sealManager.IsUnsealed() {
		log.Fatalf("Vault failed to unseal")
	}
	log.Println("Vault is UNSEALED")

	masterKey, err := sealManager.MasterKey()
	if err != nil {
		log.Fatalf("Failed to get master key from unsealed vault: %v", err)
	}
	cryptoEngine, err := crypto.NewAESGCM(masterKey)
	if err != nil {
		log.Fatalf("Failed to initialize crypto engine: %v", err)
	}

	serverConfig := server.Config{
		Storage: store,
		Seal:    sealManager,
		Crypto:  cryptoEngine,
	}

	grpcServer, err := server.NewGRPCServer(&serverConfig)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	log.Println("Starting gRPC server")
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("Failed to listen on port 8000: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on :8000")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("Shutting down gRPC server")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped")
}
