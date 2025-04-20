package core

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"log"
	"sync"
	"time"

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

// electProposer selects a proposer deterministically based on TxID
func electProposer(txID string, totalValidators int) int {
	h := fnv.New32a()
	h.Write([]byte(txID))
	return int(h.Sum32()) % totalValidators
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

func (v *Validator) StartConsensus(ctx context.Context, blockchain *Blockchain) {
	log.Printf("Validator %d starting consensus process", v.ID)

	go v.handleTransactions(ctx, blockchain)

	go v.handleVotes(ctx, blockchain)
}

func (v *Validator) handleTransactions(ctx context.Context, blockchain *Blockchain) {
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
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Validator %d transaction handler exiting", v.ID)
			return
		default:
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
				vote := VoteMessage{
					TxID:        transaction.TxID,
					ValidatorID: v.ID,
					Timestamp:   time.Now().Unix(),
					Approved:    true,
				}
				v.broadcastVote(vote)
			} else {
				log.Printf("Validator %d rejected transaction %s", v.ID, transaction.TxID)
			}

			select {
			case <-time.After(2 * time.Second):
			case <-ctx.Done():
				log.Printf("Validator %d transaction handler exiting after sleep", v.ID)
				return
			}
		}

		time.Sleep(4 * time.Second)
	}
}

type VoteMessage struct {
	TxID        string
	ValidatorID int
	Timestamp   int64
	Approved    bool
}

func (v *Validator) broadcastVote(vote VoteMessage) {
	voteData, err := json.Marshal(vote)
	if err != nil {
		log.Printf("Validator %d error marshaling vote: %v", v.ID, err)
		return
	}

	ctx := context.Background()
	if err := v.Node.VoteTopic.Publish(ctx, voteData); err != nil {
		log.Printf("Validator %d error publishing vote: %v", v.ID, err)
	} else {
		log.Printf("Validator %d broadcast vote for transaction %s", v.ID, vote.TxID)
	}
}

func (v *Validator) handleVotes(ctx context.Context, blockchain *Blockchain) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Validator %d vote handler exiting", v.ID)
			return
		default:
			msg, err := v.Node.VoteSub.Next(ctx)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			var vote VoteMessage
			if err := json.Unmarshal(msg.Data, &vote); err != nil {
				var oldVote map[string]string
				if err2 := json.Unmarshal(msg.Data, &oldVote); err2 != nil {
					log.Printf("Validator %d received invalid vote format: %v", v.ID, err)
				}
				continue
			}

			log.Printf("Validator %d received vote for transaction %s from validator %d",
				v.ID, vote.TxID, vote.ValidatorID)

			blockchain.mu.Lock()
			if blockchain.VoteCount[vote.TxID] == nil {
				blockchain.VoteCount[vote.TxID] = make(map[int]bool)
			}
			blockchain.VoteCount[vote.TxID][vote.ValidatorID] = vote.Approved
			voteCount := len(blockchain.VoteCount[vote.TxID])
			blockchain.mu.Unlock()

			if voteCount == 5 {
				blockchain.mu.Lock()
				approvals := 0
				for _, approved := range blockchain.VoteCount[vote.TxID] {
					if approved {
						approvals++
					}
				}
				blockchain.mu.Unlock()

				if approvals >= 4 {
					tx := v.findTransactionByID(vote.TxID)
					if tx != nil {
						if RegisterLicense(*tx, blockchain) {
							proposerID := electProposer(vote.TxID, 5)
							if v.ID == proposerID {
								tx.IsValidated = true
								tx.ValidatorID = v.ID
								blockchain.AddTransaction(*tx)
								v.broadcastBlockchainUpdate(blockchain.Blocks[len(blockchain.Blocks)-1])
							} else {
								log.Printf("Validator %d: Not proposer for transaction %s", v.ID, vote.TxID)
							}
							v.Mempool.RemoveTransaction(vote.TxID)
						}
					} else {
						log.Printf("Validator %d: Transaction %s not found in mempool", v.ID, vote.TxID)
					}
				} else {
					log.Printf("Transaction %s rejected: only %d approvals", vote.TxID, approvals)
					v.Mempool.RemoveTransaction(vote.TxID)
				}
			}

			select {
			case <-time.After(4 * time.Second):
			case <-ctx.Done():
				log.Printf("Validator %d vote handler exiting after sleep", v.ID)
				return
			}
		}

		time.Sleep(4 * time.Second)
	}
}

// Broadcast blockchain update to all nodes
func (v *Validator) broadcastBlockchainUpdate(block *Block) {
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
	if tx.Owner == "" || tx.AssetHash == "" {
		return false
	}

	// Verify the digital signature
	return VerifyTransaction(tx)
}
