package core

import (
	"context"
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
	Host   host.Host
	PubSub *pubsub.PubSub
	Topic  *pubsub.Topic
	Sub    *pubsub.Subscription
}

type DiscoveryNotifee struct{}

// HandlePeerFound is called when a peer is discovered
func (d *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Println("Discovered new peer:", pi)
}

func NewNode(ctx context.Context, topicName string) (*Node, error) {
	// Create a new libp2p host
	h, err := libp2p.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Create a pubsub instance
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Join the blockchain topic
	topic, err := ps.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	// Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Enable peer discovery using mDNS
	notifee := &DiscoveryNotifee{}
	service := mdns.NewMdnsService(h, "blockchain-network", notifee)
	if err := service.Start(); err != nil {
		log.Println("Failed to start mDNS:", err)
	}

	// Return the node
	return &Node{
		Host:   h,
		PubSub: ps,
		Topic:  topic,
		Sub:    sub,
	}, nil
}
