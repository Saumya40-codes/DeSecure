package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// as wiki states
/*
*  ECDSA, or Elliptic Curve Digital Signature Algorithm, is a cryptographic algorithm used to create digital signatures,
*  ensuring the authenticity and integrity of data, and is based on elliptic curve cryptography (ECC).
 */

func GenerateKeyPair() (*ecdsa.PrivateKey, string) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Error generating key pair: ", err)
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
