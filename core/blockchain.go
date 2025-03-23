package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Data      string
	PrevHash  string
	Hash      string
	Validator string
}

func calculateHash(block Block) string {
	record := fmt.Sprintf("%d%s%s%s%s", block.Index, block.Timestamp, block.Data, block.PrevHash, block.Validator)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func CreateGenesisBlock() *Block {
	genesisBlock := &Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
	}

	genesisBlock.Hash = calculateHash(*genesisBlock)
	return genesisBlock
}

func CreateBlock(prevBlock Block, data string) *Block {
	newBlock := &Block{
		Index:     prevBlock.Index + 1,
		Timestamp: time.Now().String(),
		Data:      data,
		PrevHash:  prevBlock.Hash,
	}

	newBlock.Hash = calculateHash(*newBlock)
	return newBlock
}
