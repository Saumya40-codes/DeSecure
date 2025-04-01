package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

type Block struct {
	Index       int
	Timestamp   string
	Transaction []LicenseTransaction
	PrevHash    string
	Hash        string
	Validator   string
}

type Blockchain struct {
	Blocks    []*Block
	VoteCount map[string]int
	mu        sync.Mutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:    []*Block{},
		VoteCount: make(map[string]int),
	}
}

func (bc *Blockchain) AddTransaction(tx LicenseTransaction) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Track validator votes
	bc.VoteCount[tx.TxID]++

	if bc.VoteCount[tx.TxID] >= 4 { // At least 4/5 validators approve
		prevBlock := bc.Blocks[len(bc.Blocks)-1]
		newBlock := CreateBlock(*prevBlock, []LicenseTransaction{tx})

		bc.Blocks = append(bc.Blocks, newBlock)
		log.Println("Block added with consensus:", newBlock.Hash)
	} else {
		log.Println("Transaction pending consensus:", tx.TxID)
	}
}

func (bc *Blockchain) ProcessVote(voteMsg []byte) {
	var vote map[string]string
	if err := json.Unmarshal(voteMsg, &vote); err != nil {
		log.Println("Invalid vote format:", err)
		return
	}

	txID := vote["txID"]
	bc.mu.Lock()
	bc.VoteCount[txID]++
	bc.mu.Unlock()
}

func calculateHash(block Block) string {
	txData, _ := json.Marshal(block.Transaction)
	record := fmt.Sprintf("%d%s%s%s%s", block.Index, block.Timestamp, txData, block.PrevHash, block.Validator)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func CreateGenesisBlock() *Block {
	genesisBlock := &Block{
		Index:       0,
		Timestamp:   time.Now().String(),
		Transaction: []LicenseTransaction{},
		PrevHash:    "",
	}

	genesisBlock.Hash = calculateHash(*genesisBlock)
	return genesisBlock
}

func CreateBlock(prevBlock Block, transactions []LicenseTransaction) *Block {
	newBlock := &Block{
		Index:       prevBlock.Index + 1,
		Timestamp:   time.Now().String(),
		Transaction: transactions,
		PrevHash:    prevBlock.Hash,
	}

	newBlock.Hash = calculateHash(*newBlock)
	return newBlock
}
