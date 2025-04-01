package core

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

type Validator struct {
	ID         int
	Node       *Node
	VotePool   map[string]int // Track votes per transaction
	mu         sync.Mutex
	PublicKey  string      // The validator's public key
	PrivateKey interface{} // The validator's private key
}

func NewValidator(id int, node *Node, publicKey string, privateKey interface{}) *Validator {
	return &Validator{
		ID:         id,
		Node:       node,
		VotePool:   make(map[string]int),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

func (v *Validator) StartConsensus(blockchain *Blockchain) {
	log.Printf("Validator %d starting consensus process", v.ID)

	// Start a goroutine to handle transaction messages
	go v.handleTransactions(blockchain)

	// Start a goroutine to handle vote messages
	go v.handleVotes(blockchain)
}

func (v *Validator) handleTransactions(blockchain *Blockchain) {
	ctx := context.Background()
	txSub, err := v.Node.PubSub.Subscribe("transactions")
	if err != nil {
		log.Printf("Validator %d failed to subscribe to transactions: %v", v.ID, err)
		return
	}

	for {
		msg, err := txSub.Next(ctx)
		if err != nil {
			log.Printf("Validator %d error reading transaction: %v", v.ID, err)
			continue
		}

		// Skip messages from self
		if msg.ReceivedFrom == v.Node.Host.ID() {
			continue
		}

		var transaction LicenseTransaction
		if err := json.Unmarshal(msg.Data, &transaction); err != nil {
			log.Printf("Validator %d received invalid transaction format: %v", v.ID, err)
			continue
		}

		log.Printf("Validator %d received transaction %s", v.ID, transaction.TxID)

		// Validate the transaction
		if ValidateTransaction(transaction) {
			log.Printf("Validator %d approved transaction %s", v.ID, transaction.TxID)

			// Create and broadcast vote
			vote := VoteMessage{
				TxID:        transaction.TxID,
				ValidatorID: v.ID,
				Timestamp:   time.Now().Unix(),
				Approved:    true,
				Signature:   "", // Should actually sign the vote
			}

			v.broadcastVote(vote)
		} else {
			log.Printf("Validator %d rejected transaction %s", v.ID, transaction.TxID)
		}
	}
}

type VoteMessage struct {
	TxID        string
	ValidatorID int
	Timestamp   int64
	Approved    bool
	Signature   string
}

func (v *Validator) broadcastVote(vote VoteMessage) {
	voteData, err := json.Marshal(vote)
	if err != nil {
		log.Printf("Validator %d error marshaling vote: %v", v.ID, err)
		return
	}

	ctx := context.Background()
	if err := v.Node.Topic.Publish(ctx, voteData); err != nil {
		log.Printf("Validator %d error publishing vote: %v", v.ID, err)
	} else {
		log.Printf("Validator %d broadcast vote for transaction %s", v.ID, vote.TxID)
	}
}

func (v *Validator) handleVotes(blockchain *Blockchain) {
	ctx := context.Background()
	for {
		msg, err := v.Node.Sub.Next(ctx)
		if err != nil {
			log.Printf("Validator %d error reading vote: %v", v.ID, err)
			continue
		}

		var vote VoteMessage
		if err := json.Unmarshal(msg.Data, &vote); err != nil {
			// Try the old format first
			var oldVote map[string]string
			if err2 := json.Unmarshal(msg.Data, &oldVote); err2 == nil {
				// Process old vote format
				blockchain.ProcessVote(msg.Data)
			} else {
				log.Printf("Validator %d received invalid vote format: %v", v.ID, err)
			}
			continue
		}

		log.Printf("Validator %d received vote for transaction %s from validator %d",
			v.ID, vote.TxID, vote.ValidatorID)

		// Process the vote through the blockchain
		blockchain.mu.Lock()
		blockchain.VoteCount[vote.TxID]++
		blockchain.mu.Unlock()
	}
}

func ValidateTransaction(tx LicenseTransaction) bool {
	// Basic validation
	if tx.Owner == "" || tx.AssetHash == "" {
		return false
	}

	// Verify the digital signature
	return VerifyTransaction(tx)
}

func ListenForTransactions(node *Node, blockchain *Blockchain, db *storage.DB) {
	ctx := context.Background()
	mempool := NewMempool()

	for {
		msg, err := node.Sub.Next(ctx)
		if err != nil {
			log.Println("Error reading from topic:", err)
			continue
		}

		// First try to parse as transaction
		var tx LicenseTransaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			log.Println("New transaction received:", tx.TxID)

			// Add to mempool
			mempool.AddTransaction(tx)

			// Broadcast to blockchain network
			node.BroadcastTransaction(tx)

			continue
		}

		// If not a transaction, try to parse as vote
		var vote map[string]string
		if err := json.Unmarshal(msg.Data, &vote); err == nil {
			txID, ok := vote["txID"]
			if ok {
				log.Println("Vote received for transaction:", txID)
				blockchain.ProcessVote(msg.Data)
			}
		}
	}
}

