package main

import (
	"fmt"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
)

var blockchain []*core.Block

func main() {
	// 🔹 Generate two user key pairs
	privKey1, pubKey1 := core.GenerateKeyPair()
	privKey2, pubKey2 := core.GenerateKeyPair()

	// 🔹 Create & sign a license transaction (User 1)
	tx1 := core.LicenseTransaction{
		Owner:     pubKey1,
		AssetHash: "Qm12345abcdef", // Example IPFS hash for User 1's content
		License:   "view",
	}
	tx1.Signature = core.SignTransaction(privKey1, tx1)

	// 🔹 Create & sign another license transaction (User 2)
	tx2 := core.LicenseTransaction{
		Owner:     pubKey2,
		AssetHash: "Qm67890xyzuvw", // Example IPFS hash for User 2's content
		License:   "download",
	}
	tx2.Signature = core.SignTransaction(privKey2, tx2)

	// 🔹 Verify both transactions before adding to blockchain
	if core.VerifyTransaction(tx1) && core.VerifyTransaction(tx2) {
		fmt.Println("Both transactions are valid!")
	} else {
		fmt.Println("One or more transactions are invalid!")
	}

	// 🔹 Create Genesis Block
	genesisBlock := core.CreateGenesisBlock()
	blockchain = append(blockchain, genesisBlock)

	// 🔹 Create Block 1 with both transactions
	newBlock := core.CreateBlock(*blockchain[len(blockchain)-1], []core.LicenseTransaction{tx1, tx2})
	blockchain = append(blockchain, newBlock)

	// 🔹 Print Blockchain
	for _, block := range blockchain {
		fmt.Printf("\nIndex: %d\nTimestamp: %s\nPrevHash: %s\nHash: %s\nTransactions:\n",
			block.Index, block.Timestamp, block.PrevHash, block.Hash)

		for _, tx := range block.Transaction {
			fmt.Printf("  - Owner: %s\n  - Asset: %s\n  - License: %s\n  - Signature: %s\n",
				tx.Owner, tx.AssetHash, tx.License, tx.Signature)
		}
	}
}

