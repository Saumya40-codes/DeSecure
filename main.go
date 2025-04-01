package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/cmd"
	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

const (
	TopicName   = "drm-consensus"
	DataDirPath = "./data"
)

func main() {
	if len(os.Args) > 1 {
		cmd.Execute()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down...")
		cancel()
		os.Exit(0)
	}()

	log.Println("Opening database connection...")
	os.MkdirAll(DataDirPath, 0o700)
	db := storage.OpenDB(DataDirPath)
	defer db.CloseDB()

	log.Println("Creating P2P node...")
	node, err := core.NewNode(ctx, TopicName)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}

	log.Printf("Node created with ID: %s", node.Host.ID().String())
	log.Printf("Node listening on: %v", node.Host.Addrs())

	log.Println("Initializing blockchain...")
	blockchain := core.NewBlockchain(db)
	log.Printf("Blockchain initialized with %d blocks", len(blockchain.Blocks))

	log.Println("Starting transaction listener...")
	go core.ListenForTransactions(node, blockchain, db)

	numValidators := 5

	log.Printf("Starting %d validators...", numValidators)
	for i := 0; i < numValidators; i++ {
		go func(id int) {
			privKey, pubKey := core.GenerateKeyPair()
			log.Printf("Validator %d initialized with public key: %s", id, pubKey[:16]+"...")

			validator := core.NewValidator(id, node, pubKey, privKey)
			validator.StartConsensus(blockchain)
		}(i)
	}

	select {}
}

