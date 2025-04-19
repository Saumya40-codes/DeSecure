package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
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
}

// Global License Registry
var licenseRegistry = struct {
	sync.Mutex
	licenses map[string]LicenseTransaction // AssetHash -> LicenseTransaction
}{licenses: make(map[string]LicenseTransaction)}

// Generate a new key pai
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
func RegisterLicense(transaction LicenseTransaction) bool {
	if !VerifyTransaction(transaction) {
		fmt.Println("Invalid license transaction")
		return false
	}

	licenseRegistry.Lock()
	defer licenseRegistry.Unlock()

	if _, exists := licenseRegistry.licenses[transaction.AssetHash]; exists {
		fmt.Println("License already exists for asset:", transaction.AssetHash)
		return false
	}

	licenseRegistry.licenses[transaction.AssetHash] = transaction
	fmt.Println("License registered:", transaction.AssetHash, "Owner:", transaction.Owner)
	return true
}

// Check if a user has a valid, unexpired license
func HasValidLicense(user string, assetHash string) bool {
	licenseRegistry.Lock()
	defer licenseRegistry.Unlock()

	license, exists := licenseRegistry.licenses[assetHash]
	if !exists {
		return false
	}

	// Check if the user is the owner or the designated licensee
	if license.Owner != user && (license.Licensee != "" && license.Licensee != user) {
		return false
	}

	// Check if the license is expired
	if license.Expiry > 0 && time.Now().Unix() > license.Expiry {
		fmt.Println("License expired for asset:", assetHash)
		return false
	}

	return true
}
