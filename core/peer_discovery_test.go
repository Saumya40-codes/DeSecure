package core

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestPeerDiscovery(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create first node
	node1, err := NewNode(ctx, "test-network", false)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	log.Printf("Node1 created with ID: %s", node1.Host.ID().String())

	// Create second node
	node2, err := NewNode(ctx, "test-network", false)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	log.Printf("Node2 created with ID: %s", node2.Host.ID().String())

	// Create third node
	node3, err := NewNode(ctx, "test-network", false)
	if err != nil {
		t.Fatalf("Failed to create node3: %v", err)
	}
	log.Printf("Node3 created with ID: %s", node3.Host.ID().String())

	// Wait for peer discovery to happen
	time.Sleep(10 * time.Second)

	// Check if nodes discovered each other
	peers1 := node1.Host.Network().Peers()
	peers2 := node2.Host.Network().Peers()
	peers3 := node3.Host.Network().Peers()

	log.Printf("Node1 peers: %v", peers1)
	log.Printf("Node2 peers: %v", peers2)
	log.Printf("Node3 peers: %v", peers3)

	// Verify that each node has discovered the others
	if len(peers1) < 2 {
		t.Errorf("Node1 should have discovered at least 2 peers, got %d", len(peers1))
	}
	if len(peers2) < 2 {
		t.Errorf("Node2 should have discovered at least 2 peers, got %d", len(peers2))
	}
	if len(peers3) < 2 {
		t.Errorf("Node3 should have discovered at least 2 peers, got %d", len(peers3))
	}
} 