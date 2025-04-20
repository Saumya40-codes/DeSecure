package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sync"
)

// LicenseTransaction struct with additional fields
type LicenseTransaction struct {
	TxID        string // Unique transaction ID
	Owner       string // Public key of the owner
	AssetHash   string // Unique identifier for the asset
	License     string // License type (e.g., view, download)
	Signature   string // Digital signature for authenticity
	ValidatorID int    // Id of the validator who validated the transaction
	Metadata    string // JSON Metadata (Title, Description, Category)
	Timestamp   int64  // Unix timestamp of the transaction
	Expiry      int64  // Unix timestamp for expiration (optional, 0 if no expiry)
	Licensee    string // Public key of the license recipient (if applicable)
	IsValidated bool   // Whether the transaction has been validated
	Nonce       uint64 // We can use this for transaction replay protection
	TxType      string // Transaction type: "upload", "purchase", etc.
	// Price       float64 // a hypothetical blockchain, no we dont need price
}

// Global License Registry
var licenseRegistry = struct {
	sync.Mutex
	licenses map[string]LicenseTransaction // AssetHash -> LicenseTransaction
}{licenses: make(map[string]LicenseTransaction)}

// Generate a new key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, string) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Error generating key pair:", err)
		return nil, ""
	}

	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
	pubKeyHex := hex.EncodeToString(pubKey)

	return privKey, pubKeyHex
}

// Generate a unique transaction ID
func GenerateTransactionID(transaction LicenseTransaction) string {
	data := transaction.Owner + transaction.AssetHash + transaction.License + fmt.Sprintf("%d", transaction.Timestamp)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Sign the license transaction
func SignTransaction(privKey *ecdsa.PrivateKey, transaction *LicenseTransaction) string {
	data := transaction.Owner + transaction.AssetHash + transaction.License + transaction.TxID + fmt.Sprintf("%d", transaction.Timestamp)
	hash := sha256.Sum256([]byte(data))

	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		fmt.Println("Error signing transaction:", err)
		return ""
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature)
}

// Verify the transaction signature
func VerifyTransaction(transaction LicenseTransaction) bool {
	pubKeyBytes, _ := hex.DecodeString(transaction.Owner)
	x, y := new(big.Int).SetBytes(pubKeyBytes[:32]), new(big.Int).SetBytes(pubKeyBytes[32:])
	pubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

	data := transaction.Owner + transaction.AssetHash + transaction.License + transaction.TxID + fmt.Sprintf("%d", transaction.Timestamp)
	hash := sha256.Sum256([]byte(data))

	signBytes, _ := hex.DecodeString(transaction.Signature)
	r, s := new(big.Int).SetBytes(signBytes[:32]), new(big.Int).SetBytes(signBytes[32:])

	return ecdsa.Verify(&pubKey, hash[:], r, s)
}

// Register a new license
func RegisterLicense(transaction LicenseTransaction, bc *Blockchain) bool {
	if !VerifyTransaction(transaction) {
		fmt.Println("Invalid license transaction")
		return false
	}

	licenseRegistry.Lock()
	defer licenseRegistry.Unlock()

	for _, block := range bc.Blocks {
		for _, existingTx := range block.Transaction {
			// Check for proper nonce sequence
			if existingTx.Owner == transaction.Owner && existingTx.Nonce >= transaction.Nonce {
				log.Println("Invalid nonce")
				return false
			}
		}
	}

	if transaction.TxType == "upload" {
		if val, _ := bc.db.Load(transaction.AssetHash); val != nil {
			log.Println("License already exists for asset:", transaction.AssetHash)
			return false
		}

		licenseRegistry.licenses[transaction.AssetHash] = transaction
		jsonTransac, err := json.Marshal(transaction)
		if err != nil {
			log.Fatal("error occured")
		}

		bc.db.Save(transaction.AssetHash, jsonTransac)
		fmt.Println("License registered:", transaction.AssetHash, "Owner:", transaction.Owner)
	} else {
		if val, _ := bc.db.Load(transaction.AssetHash); val == nil {
			log.Fatal("The asset doesn't exists")
		}

		for _, block := range bc.Blocks {
			for _, existingTx := range block.Transaction {
				if existingTx.AssetHash == transaction.AssetHash && existingTx.Owner != transaction.Owner {
					log.Fatal("Invalid transaction")
				}
			}
		}
	}
	return true
}

// Check if a user has a valid, unexpired license
func HasValidLicense(assetHash string, bc *Blockchain, pubKey string) bool {
	licenseRegistry.Lock()
	defer licenseRegistry.Unlock()

	for _, block := range bc.Blocks {
		for _, existingTx := range block.Transaction {
			if existingTx.AssetHash == assetHash && (existingTx.License == pubKey || existingTx.Owner == pubKey) {
				return true
			}
		}
	}

	return false
}
