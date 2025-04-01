package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/secretbox"
)

const keyDir = "./keys/"

var secretKey = getSecretKey()

const secretKeyFile = ".secret_key"

// Generate or load the encryption key
func getSecretKey() [32]byte {
	keyPath := filepath.Join(keyDir, secretKeyFile)

	if _, err := os.Stat(keyPath); err == nil {
		data, _ := os.ReadFile(keyPath)
		var storedKey [32]byte
		copy(storedKey[:], data)
		return storedKey
	}

	var newKey [32]byte
	_, err := rand.Read(newKey[:])
	if err != nil {
		fmt.Println("Error generating encryption key:", err)
		os.Exit(1)
	}

	os.MkdirAll(keyDir, 0o700)
	_ = os.WriteFile(keyPath, newKey[:], 0o600)

	return newKey
}

var (
	filePath    string
	title       string
	description string
	category    string
	license     string
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to IPFS and register license transaction",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a node to broadcast the transaction
		ctx := context.Background()
		node, err := core.NewNode(ctx, "drm-consensus")
		if err != nil {
			fmt.Println("Error creating P2P node:", err)
			return
		}

		privKey, pubKey := ensureKeyPair()

		cid, err := storage.UploadtoIPFS(filePath)
		if err != nil {
			fmt.Println("Error uploading file to IPFS:", err)
			return
		}
		fmt.Println("✅ File uploaded to IPFS, CID:", cid)

		assetHash := cid

		// Create metadata object
		metadata := map[string]string{
			"Title":       title,
			"Description": description,
			"Category":    category,
		}

		metadataJSON, _ := json.Marshal(metadata)

		// Create transaction with all required fields
		transaction := core.LicenseTransaction{
			Owner:       pubKey,
			AssetHash:   assetHash,
			License:     license,
			Metadata:    string(metadataJSON),
			Timestamp:   time.Now().Unix(),
			IsValidated: false,
		}

		// Generate transaction ID
		transaction.TxID = core.GenerateTransactionID(transaction)

		// Sign the transaction
		transaction.Signature = core.SignTransaction(privKey, &transaction)

		// Instead of accessing the local registry, directly broadcast to network
		fmt.Println("🌐 Broadcasting transaction to network for validation...")
		node.BroadcastTransaction(transaction)
		fmt.Println("✅ Transaction broadcast complete! TxID:", transaction.TxID)
		fmt.Println("ℹ️ Your transaction will be validated by the network and added to the blockchain.")
		fmt.Println("ℹ️ You can check its status later using the blockchain command.")

		// Wait a brief moment to ensure message is sent before the program exits
		time.Sleep(2 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the file to upload")
	uploadCmd.Flags().StringVarP(&title, "title", "t", "Untitled", "Title of the asset")
	uploadCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the asset")
	uploadCmd.Flags().StringVarP(&category, "category", "c", "Uncategorized", "Category of the asset")
	uploadCmd.Flags().StringVarP(&license, "license", "l", "view", "License type (view, download, etc.)")
	uploadCmd.MarkFlagRequired("file")
}

// Ensure key pair exists; otherwise, generate one
func ensureKeyPair() (*ecdsa.PrivateKey, string) {
	privPath, _ := keyPaths()
	if _, err := os.Stat(privPath); err == nil {
		return loadKeyPair()
	}
	return generateAndSaveKeyPair()
}

// Generate and save new key pair
func generateAndSaveKeyPair() (*ecdsa.PrivateKey, string) {
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubKey := hex.EncodeToString(append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...))
	privBytes, _ := x509.MarshalECPrivateKey(privKey)

	var nonce [24]byte
	encrypted := secretbox.Seal(nonce[:], privBytes, &nonce, &secretKey)

	os.MkdirAll(keyDir, 0o700)
	_ = os.WriteFile(filepath.Join(keyDir, ".private_key"), encrypted, 0o600)
	_ = os.WriteFile(filepath.Join(keyDir, ".public_key"), []byte(pubKey), 0o644)

	return privKey, pubKey
}

// Load existing key pair
func loadKeyPair() (*ecdsa.PrivateKey, string) {
	privPath, pubPath := keyPaths()
	encrypted, _ := os.ReadFile(privPath)
	pubKey, _ := os.ReadFile(pubPath)

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])
	privBytes, _ := secretbox.Open(nil, encrypted[24:], &nonce, &secretKey)

	privKey, _ := x509.ParseECPrivateKey(privBytes)
	return privKey, string(pubKey)
}

// Get key file paths using glob-style logic
func keyPaths() (string, string) {
	return filepath.Join(keyDir, ".private_key"), filepath.Join(keyDir, ".public_key")
}
