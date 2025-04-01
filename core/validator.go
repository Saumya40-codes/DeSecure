package core

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

type Validator struct {
	ID       int
	Node     *Node
	VotePool map[string]int // Track votes per transaction
	mu       sync.Mutex
}

func NewValidator(id int, node *Node) *Validator {
	return &Validator{
		ID:       id,
		Node:     node,
		VotePool: make(map[string]int),
	}
}

func (v *Validator) StartConsensus(blockchain *Blockchain) {
	for {
		msg, err := v.Node.Sub.Next(context.Background())
		if err != nil {
			log.Println("Error reading from topic:", err)
			continue
		}

		var transaction LicenseTransaction
		if err := json.Unmarshal(msg.Data, &transaction); err != nil {
			log.Println("Invalid transaction format:", err)
			continue
		}

		if ValidateTransaction(transaction) {
			v.mu.Lock()
			v.VotePool[transaction.TxID]++
			v.mu.Unlock()

			log.Println("Validator", v.ID, "approved transaction", transaction.TxID)

			// Broadcast the vote
			voteMsg, _ := json.Marshal(map[string]string{
				"txID":      transaction.TxID,
				"validator": string(v.ID),
			})
			v.Node.Topic.Publish(context.Background(), voteMsg)
		}
	}
}

func ValidateTransaction(tx LicenseTransaction) bool {
	return tx.Owner != "" && tx.AssetHash != ""
}

func ListenForTransactions(node *Node, blockchain *Blockchain, db *storage.DB) {
	for {
		msg, err := node.Sub.Next(context.Background())
		if err != nil {
			log.Println("Error reading from topic:", err)
			continue
		}
		var tx LicenseTransaction
		if err := json.Unmarshal(msg.Data, &tx); err != nil {
			log.Println("Invalid transaction data:", err)
			continue
		}
		log.Println("New transaction received:", tx.TxID)
		blockchain.AddTransaction(tx)
	}
}
