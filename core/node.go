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
	Host   host.Host
	PubSub *pubsub.PubSub
	Topic  *pubsub.Topic
	Sub    *pubsub.Subscription
}

func (n *Node) BroadcastTransaction(tx LicenseTransaction) {
	txData, _ := json.Marshal(tx)
	if err := n.Topic.Publish(context.Background(), txData); err != nil {
		log.Println("Error broadcasting transaction:", err)
	}
	log.Println("Broadcasted!!")
}

type DiscoveryNotifee struct {
	host host.Host
}

func (d *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Println("Discovered new peer:", pi.ID)

	if err := d.host.Connect(context.Background(), pi); err != nil {
		log.Println("Error connecting to discovered peer:", err)
	} else {
		log.Println("Connected to discovered peer:", pi.ID)
	}
}

func NewNode(ctx context.Context, topicName string) (*Node, error) {
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

	// Enable peer discovery using mDNS
	notifee := &DiscoveryNotifee{host: h}
	service := mdns.NewMdnsService(h, "blockchain-network", notifee)
	if err := service.Start(); err != nil {
		log.Println("Failed to start mDNS:", err)
	}

	return &Node{
		Host:   h,
		PubSub: ps,
		Topic:  topic,
		Sub:    sub,
	}, nil
}
