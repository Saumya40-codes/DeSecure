package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type Validator struct {
	ID         int
	Node       *Node
	VotePool   map[string]int // Track votes per transaction
	mu         sync.Mutex
	PublicKey  string      // The validator's public key
	PrivateKey interface{} // The validator's private key
	Mempool    *Mempool    // Add this field
}

var (
	topicJoin     *pubsub.Topic
	topicJoinOnce sync.Once
)

// Find a transaction by ID from the mempool
func (v *Validator) findTransactionByID(txID string) *LicenseTransaction {
	if v.Mempool == nil {
		log.Printf("Validator %d: Mempool is not initialized", v.ID)
		return nil
	}

	return v.Mempool.GetTransactionByID(txID)
}

func NewValidator(id int, node *Node, publicKey string, privateKey interface{}, mempool *Mempool) *Validator {
	return &Validator{
		ID:         id,
		Node:       node,
		VotePool:   make(map[string]int),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		Mempool:    mempool,
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

	topicJoinOnce.Do(func() {
		var err error
		topicJoin, err = v.Node.PubSub.Join("transactions")
		if err != nil {
			log.Printf("Validator %d failed to join transactions topic: %v", v.ID, err)
			return
		}
	})

	if topicJoin == nil {
		log.Printf("Validator %d: Topic join failed earlier", v.ID)
		return
	}

	txSub, err := topicJoin.Subscribe()
	if err != nil {
		log.Printf("Validator %d failed to subscribe to transactions: %v", v.ID, err)
		return
	} else {
		log.Println("Topic subscribed")
	}

	for {
		msg, err := txSub.Next(ctx)
		if err != nil {
			log.Printf("Validator %d error reading transaction: %v", v.ID, err)
			continue
		}

		if msg.ReceivedFrom == v.Node.Host.ID() {
			continue
		}

		var transaction LicenseTransaction
		if err := json.Unmarshal(msg.Data, &transaction); err != nil {
			log.Printf("Validator %d received invalid transaction format: %v", v.ID, err)
			continue
		}

		v.Mempool.AddTransaction(transaction)

		log.Printf("Validator %d received transaction %s", v.ID, transaction.TxID)

		if ValidateTransaction(transaction) {
			log.Printf("Validator %d approved transaction %s", v.ID, transaction.TxID)

			vote := VoteMessage{
				TxID:        transaction.TxID,
				ValidatorID: v.ID,
				Timestamp:   time.Now().Unix(),
				Approved:    true,
				Signature:   "",
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
		voteCount := blockchain.VoteCount[vote.TxID]
		blockchain.mu.Unlock()

		// Check if we have enough votes for consensus
		if voteCount >= 4 { // At least 4/5 validators approve
			// Find transaction in mempool
			tx := v.findTransactionByID(vote.TxID)
			if tx != nil {
				log.Printf("Validator %d: Consensus reached for transaction %s, adding to blockchain",
					v.ID, vote.TxID)
				blockchain.AddTransaction(*tx)

				// Remove transaction from mempool after adding to blockchain
				v.Mempool.RemoveTransaction(vote.TxID)

				// Broadcast updated blockchain state
				v.broadcastBlockchainUpdate(blockchain.Blocks[len(blockchain.Blocks)-1])
			} else {
				log.Printf("Validator %d: Transaction %s not found in mempool", v.ID, vote.TxID)
			}
		}
	}
}

// Broadcast blockchain update to all nodes
func (v *Validator) broadcastBlockchainUpdate(block *Block) {
	fmt.Println("Broadcast called")
	// Create a message that indicates this is a block update
	updateMsg := map[string]interface{}{
		"type":  "block_update",
		"block": block,
	}

	msgData, err := json.Marshal(updateMsg)
	if err != nil {
		log.Printf("Validator %d error creating block update message: %v", v.ID, err)
		return
	}

	ctx := context.Background()
	if err := v.Node.Topic.Publish(ctx, msgData); err != nil {
		log.Printf("Validator %d error broadcasting block update: %v", v.ID, err)
	} else {
		log.Printf("Validator %d broadcast block update: %s", v.ID, block.Hash)
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

func ListenForTransactions(node *Node, blockchain *Blockchain, db *storage.DB, mempool *Mempool) {
	ctx := context.Background()

	for {
		msg, err := node.Sub.Next(ctx)
		if err != nil {
			log.Println("Error reading from topic:", err)
			continue
		}

		// Try to determine if this is a block update
		var msgType struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg.Data, &msgType); err == nil && msgType.Type == "block_update" {
			// Handle block update
			var updateMsg struct {
				Type  string `json:"type"`
				Block Block  `json:"block"`
			}
			if err := json.Unmarshal(msg.Data, &updateMsg); err == nil {
				log.Println("Received block update:", updateMsg.Block.Hash)

				// Check if we already have this block
				hasBlock := false
				blockchain.mu.Lock()
				for _, b := range blockchain.Blocks {
					if b.Hash == updateMsg.Block.Hash {
						hasBlock = true
						break
					}
				}

				if !hasBlock {
					// Add the block to our blockchain
					blockCopy := updateMsg.Block // Make a copy to avoid issues with the pointer
					blockchain.Blocks = append(blockchain.Blocks, &blockCopy)
					blockchain.persistBlock(&blockCopy)

					log.Println("Added new block from network:", blockCopy.Hash)

					// Clean up mempool
					for _, tx := range blockCopy.Transaction {
						mempool.RemoveTransaction(tx.TxID)
					}
				}
				blockchain.mu.Unlock()
			}
			continue
		}

		// Try to parse as transaction
		var tx LicenseTransaction
		if err := json.Unmarshal(msg.Data, &tx); err == nil {
			log.Println("New transaction received:", tx.TxID)

			// Add to mempool
			mempool.AddTransaction(tx)

			continue
		}

		// Try to parse as vote
		var vote VoteMessage
		if err := json.Unmarshal(msg.Data, &vote); err == nil {
			log.Printf("Vote received for transaction: %s", vote.TxID)

			blockchain.mu.Lock()
			blockchain.VoteCount[vote.TxID]++
			blockchain.mu.Unlock()

			continue
		}

		// Try old vote format
		var oldVote map[string]string
		if err := json.Unmarshal(msg.Data, &oldVote); err == nil {
			txID, ok := oldVote["txID"]
			if ok {
				log.Println("Vote received for transaction (old format):", txID)
				blockchain.ProcessVote(msg.Data)
			}
		}
	}
}
