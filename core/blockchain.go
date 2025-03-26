package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	Blocks []*Block
}

func NewBlockchain() *Blockchain {
	return &Blockchain{}
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
