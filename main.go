package main

import (
	"fmt"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
)

var blockchain []*core.Block

func main() {
	fmt.Println("Creating genesis block")

	genesisBlock := core.CreateGenesisBlock()
	blockchain = append(blockchain, genesisBlock)

	fmt.Println("adding first block..")
	block1 := core.CreateBlock(*blockchain[len(blockchain)-1], "Block 1")
	blockchain = append(blockchain, block1)

	fmt.Println("adding second block...")
	block2 := core.CreateBlock(*blockchain[len(blockchain)-1], "Block 2")
	blockchain = append(blockchain, block2)

	for _, block := range blockchain {
		fmt.Printf("\nIndex: %d\nTimestamp: %s\nData: %s\nPrevHash: %s\nHash: %s\n",
			block.Index, block.Timestamp, block.Data, block.PrevHash, block.Hash)
	}
}
