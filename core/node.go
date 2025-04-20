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

	// Enable peer discovery using mDNS
	notifee := &DiscoveryNotifee{host: h}
	service := mdns.NewMdnsService(h, "blockchain-network", notifee)
	if err := service.Start(); err != nil {
		log.Println("Failed to start mDNS:", err)
	}

	newNode, err := &Node{
		Host:      h,
		PubSub:    ps,
		Topic:     topic,
		Sub:       sub,
		VoteTopic: voteTopic,
		VoteSub:   voteSub,
	}, nil

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

	return newNode, err
}
