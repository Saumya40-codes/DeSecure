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
)

type LicenseTransaction struct {
	Owner     string // Public key of the owner
	AssetHash string // Unique identifier for the asset
	License   string // License type (e.g., view, download)
	Signature string // Digital signature for authenticity
	Metadata  string // JSON Metadata (Title, Description, Category)
}

var licenseRegistry = struct {
	sync.Mutex
	licenses map[string]string // AssetHash -> Owner
}{licenses: make(map[string]string)}

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

func SignTransaction(privKey *ecdsa.PrivateKey, transaction LicenseTransaction) string {
	data := transaction.Owner + transaction.AssetHash + transaction.License
	hash := sha256.Sum256([]byte(data))

	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		fmt.Println("Error signing transaction:", err)
		return ""
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature)
}

func VerifyTransaction(transaction LicenseTransaction) bool {
	pubKeyBytes, _ := hex.DecodeString(transaction.Owner)
	x, y := new(big.Int).SetBytes(pubKeyBytes[:32]), new(big.Int).SetBytes(pubKeyBytes[32:])
	pubKey := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}

	data := transaction.Owner + transaction.AssetHash + transaction.License
	hash := sha256.Sum256([]byte(data))

	signBytes, _ := hex.DecodeString(transaction.Signature)
	r, s := new(big.Int).SetBytes(signBytes[:32]), new(big.Int).SetBytes(signBytes[32:])

	return ecdsa.Verify(&pubKey, hash[:], r, s)
}

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

	licenseRegistry.licenses[transaction.AssetHash] = transaction.Owner
	fmt.Println("License registered:", transaction.AssetHash, "Owner:", transaction.Owner)
	return true
}

func HasValidLicense(user string, assetHash string) bool {
	licenseRegistry.Lock()
	defer licenseRegistry.Unlock()

	owner, exists := licenseRegistry.licenses[assetHash]
	return exists && owner == user
}
