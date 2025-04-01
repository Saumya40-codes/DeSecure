package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

// Add storage-related constants
const (
	BlockPrefix    = "block-"
	LatestBlockKey = "latest-block"
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
	db        *storage.DB // Add DB reference
}

func NewBlockchain(db *storage.DB) *Blockchain {
	bc := &Blockchain{
		Blocks:    []*Block{},
		VoteCount: make(map[string]int),
		db:        db,
	}

	// Try to load existing blockchain from DB
	bc.loadFromDB()

	// If no blocks were loaded, create genesis block
	if len(bc.Blocks) == 0 {
		genesis := CreateGenesisBlock()
		bc.Blocks = append(bc.Blocks, genesis)
		bc.persistBlock(genesis)
	}

	return bc
}

// Load blockchain from database
func (bc *Blockchain) loadFromDB() {
	latestHashBytes, err := bc.db.Load(LatestBlockKey)
	if err != nil {
		log.Println("No blockchain found in database, starting fresh")
		return
	}

	latestHash := string(latestHashBytes)
	var currentHash = latestHash

	// Reconstruct blockchain by walking backward from latest block
	for currentHash != "" {
		blockData, err := bc.db.Load(BlockPrefix + currentHash)
		if err != nil {
			log.Println("Error loading block:", err)
			break
		}

		var block Block
		if err := json.Unmarshal(blockData, &block); err != nil {
			log.Println("Error unmarshaling block:", err)
			break
		}

		// Prepend to maintain correct order
		bc.Blocks = append([]*Block{&block}, bc.Blocks...)

		// Move to previous block
		currentHash = block.PrevHash
		if currentHash == "" {
			// Reached genesis block
			break
		}
	}

	log.Printf("Loaded %d blocks from database", len(bc.Blocks))
}

// Persist block to database
func (bc *Blockchain) persistBlock(block *Block) {
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Println("Error marshaling block:", err)
		return
	}

	// Save block by hash
	if err := bc.db.Save(BlockPrefix+block.Hash, blockData); err != nil {
		log.Println("Error saving block:", err)
		return
	}

	// Update latest block pointer
	if err := bc.db.Save(LatestBlockKey, []byte(block.Hash)); err != nil {
		log.Println("Error updating latest block:", err)
		return
	}

	log.Printf("Block %d with hash %s persisted to database", block.Index, block.Hash)
}

func (bc *Blockchain) AddTransaction(tx LicenseTransaction) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Track validator votes
	bc.VoteCount[tx.TxID]++

	if bc.VoteCount[tx.TxID] >= 4 { // At least 4/5 validators approve
		if len(bc.Blocks) == 0 {
			log.Println("Error: No blocks in blockchain")
			return
		}

		prevBlock := bc.Blocks[len(bc.Blocks)-1]
		newBlock := CreateBlock(*prevBlock, []LicenseTransaction{tx})

		bc.Blocks = append(bc.Blocks, newBlock)
		bc.persistBlock(newBlock) // Persist new block

		// Clear vote count for this transaction
		delete(bc.VoteCount, tx.TxID)

		log.Println("Block added with consensus:", newBlock.Hash)
	} else {
		log.Println("Transaction pending consensus:", tx.TxID, "Votes:", bc.VoteCount[tx.TxID])
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

	log.Println("Vote processed for transaction:", txID, "Current votes:", bc.VoteCount[txID])
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
		Validator:   "genesis",
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
		Validator:   "validator", // This should actually be set to the validator who proposed the block
	}

	newBlock.Hash = calculateHash(*newBlock)
	return newBlock
}

