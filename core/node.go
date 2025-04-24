package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// Node represents a blockchain node (uploader or validator)
type Node struct {
	Host      host.Host
	PubSub    *pubsub.PubSub
	Topic     *pubsub.Topic
	Sub       *pubsub.Subscription
	VoteTopic *pubsub.Topic
	VoteSub   *pubsub.Subscription
	// Balance   uint64  a hypothetical blockchain, no we dont need price
}

// PeerDiscoveryMessage represents a message broadcast when a new peer joins
type PeerDiscoveryMessage struct {
	Type      string `json:"type"`
	PeerID    string `json:"peer_id"`
	Addresses []string `json:"addresses"`
}

func (n *Node) BroadcastTransaction(tx LicenseTransaction) {
	txData, _ := json.Marshal(tx)
	if err := n.Topic.Publish(context.Background(), txData); err != nil {
		log.Println("Error broadcasting transaction:", err)
	}
	log.Println("Broadcasted!!")
}

// BroadcastNewPeer broadcasts a message when a new peer is discovered
func (n *Node) BroadcastNewPeer(peerID peer.ID, addresses []string) {
	msg := PeerDiscoveryMessage{
		Type:      "peer_discovery",
		PeerID:    peerID.String(),
		Addresses: addresses,
	}
	
	msgData, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling peer discovery message:", err)
		return
	}

	if err := n.Topic.Publish(context.Background(), msgData); err != nil {
		log.Println("Error broadcasting peer discovery:", err)
	} else {
		log.Printf("Broadcasted new peer discovery: %s", peerID)
	}
}

type DiscoveryNotifee struct {
	host host.Host
	node *Node
}

func (d *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Println("Discovered new peer:", pi.ID)

	if err := d.host.Connect(context.Background(), pi); err != nil {
		log.Println("Error connecting to discovered peer:", err)
	} else {
		log.Println("Connected to discovered peer:", pi.ID)
		// Broadcast the new peer to the network
		var addresses []string
		for _, addr := range pi.Addrs {
			addresses = append(addresses, addr.String())
		}
		d.node.BroadcastNewPeer(pi.ID, addresses)
	}
}

type Service interface {
	Close() error
}

func NewNode(ctx context.Context, topicName string, isValidator bool) (*Node, error) {
	h, err := libp2p.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithFloodPublish(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	var voteTopic *pubsub.Topic
	var voteSub *pubsub.Subscription
	if isValidator {
		voteTopic, err = ps.Join("vote")
		if err != nil {
			return nil, fmt.Errorf("failed to join topic: %w", err)
		}

		voteSub, err = voteTopic.Subscribe()
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
		}
	}

	newNode := &Node{
		Host:      h,
		PubSub:    ps,
		Topic:     topic,
		Sub:       sub,
		VoteTopic: voteTopic,
		VoteSub:   voteSub,
	}

	// Enable peer discovery using mDNS
	notifee := &DiscoveryNotifee{host: h, node: newNode}
	service := mdns.NewMdnsService(h, "blockchain-network", notifee)
	if err := service.Start(); err != nil {
		log.Println("Failed to start mDNS:", err)
	}

	// Start listening for peer discovery messages
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				log.Println("Error reading from topic:", err)
				continue
			}

			var discoveryMsg PeerDiscoveryMessage
			if err := json.Unmarshal(msg.Data, &discoveryMsg); err == nil && discoveryMsg.Type == "peer_discovery" {
				log.Printf("Received peer discovery message for peer: %s", discoveryMsg.PeerID)
				// Here you can add additional logic to handle the new peer
				// For example, connecting to the peer if not already connected
			}
		}
	}()

	go func(ctx context.Context, service Service, node *Node) {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, shutting down mDNS service...")
			if node.Sub != nil {
				node.Sub.Cancel()
			}
			if node.VoteSub != nil {
				node.VoteSub.Cancel()
			}
			service.Close()
		}
	}(ctx, service, newNode)

	return newNode, nil
}
