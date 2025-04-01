package main

import (
	"context"
	"log"
	"sync"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

const TopicName = "drm-consensus"

func main() {
	ctx := context.Background()

	node, err := core.NewNode(ctx, TopicName)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}
	db := storage.OpenDB("./data")
	defer db.CloseDB()

	blockchain := core.NewBlockchain()
	if len(blockchain.Blocks) == 0 {
		genesis := core.CreateGenesisBlock()
		blockchain.Blocks = append(blockchain.Blocks, genesis)
	}

	var wg sync.WaitGroup
	numValidators := 4 // Number of validator nodes
	for i := 0; i < numValidators; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			validator := core.NewValidator(id, node)
			validator.StartConsensus(blockchain)
		}(i)
	}

	go core.ListenForTransactions(node, blockchain, db)

	wg.Wait()
}
