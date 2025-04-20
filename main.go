package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/cmd"
	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

const (
	TopicName            = "drm-consensus"
	DataDirPath          = "./data"
	ValidatorDataDirPath = "./validator"
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
	}()

	log.Println("Opening database connection...")
	os.MkdirAll(DataDirPath, 0o700)
	db := storage.OpenDB(DataDirPath)
	defer db.CloseDB()

	log.Println("Creating P2P node...")
	node, err := core.NewNode(ctx, TopicName, false)
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

	wg := &sync.WaitGroup{}
	log.Printf("Starting %d validators...", numValidators)

	for i := range numValidators {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			privKey, pubKey := core.GenerateKeyPair()
			log.Printf("Validator %d initialized with public key: %s", id, pubKey[:16]+"...")

			validatorDBPath := ValidatorDataDirPath + "/validator_" + strconv.Itoa(id)
			os.MkdirAll(validatorDBPath, 0o700)
			db := storage.OpenDB(validatorDBPath)
			defer func() {
				db.CloseDB()
				log.Printf("Validator %d DB closed", id)
			}()

			node, err := core.NewNode(ctx, TopicName, true)
			if err != nil {
				log.Printf("Validator %d failed to create node: %v", id, err)
				return
			}

			validatorBlockchain := core.NewBlockchain(db)
			mempool := core.NewMempool()

			validator := core.NewValidator(id, node, pubKey, privKey, mempool)
			validator.StartConsensus(ctx, validatorBlockchain)

			<-ctx.Done()
			log.Printf("Validator %d received shutdown signal", id)
		}(i)
	}

	<-ctx.Done()
	log.Println("Main context canceled, waiting for validators to shut down...")
	wg.Wait()
	log.Println("All validators shut down cleanly.")
}
