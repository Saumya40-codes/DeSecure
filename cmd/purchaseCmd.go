package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Saumya40-codes/DeSecure/core"
	storage "github.com/Saumya40-codes/DeSecure/pkg"
	"github.com/spf13/cobra"
)

var (
	assetID string
	price   float64
)

var purchaseCmd = &cobra.Command{
	Use:   "purchase",
	Short: "Purchase a license for an asset on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		node, err := core.NewNode(ctx, "transactions", false)
		if err != nil {
			fmt.Println("Error creating P2P node:", err)
			return
		}

		privKey, pubKey := ensureKeyPair()

		// Open database to get asset information
		db := storage.OpenDB("./data")
		defer db.CloseDB()

		var originalTx core.LicenseTransaction
		var assetData []byte

		bc := core.NewBlockchain(db)

		for _, blocks := range bc.Blocks {
			for _, tx := range blocks.Transaction {
				if tx.AssetHash == assetID {
					assetData, err = json.Marshal(tx)
					if err == nil {
						db.Save(assetID, assetData)
					} else {
						db.CloseDB()
						log.Fatal("Error occurred")
						fmt.Println(err)
					}
				}
			}
		}

		if err := json.Unmarshal(assetData, &originalTx); err != nil {
			fmt.Println("❌ Error decoding asset data:", err)
			return
		}

		// Create purchase transaction
		purchaseTx := core.LicenseTransaction{
			Owner:       originalTx.Owner,    // Original owner
			Licensee:    pubKey,              // New licensee (buyer)
			AssetHash:   assetID,             // Asset being purchased
			License:     originalTx.License,  // Keep same license type
			Metadata:    originalTx.Metadata, // Keep same metadata
			Timestamp:   time.Now().Unix(),
			IsValidated: false,
			TxType:      "purchase", // Mark as purchase transaction
		}

		purchaseTx.TxID = core.GenerateTransactionID(purchaseTx)
		purchaseTx.Signature = core.SignTransaction(privKey, &purchaseTx)

		fmt.Println("🌐 Broadcasting purchase transaction to network for validation...")
		node.BroadcastTransaction(purchaseTx)

		fmt.Println("✅ Purchase request broadcast complete! TxID:", purchaseTx.TxID)
		fmt.Println("ℹ️ Your purchase will be validated by the network and added to the blockchain.")
		fmt.Println("ℹ️ You can check its status later using the blockchain command.")

		time.Sleep(2 * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(purchaseCmd)
	purchaseCmd.Flags().StringVarP(&assetID, "asset", "a", "", "Asset ID/hash to purchase")
	purchaseCmd.Flags().Float64VarP(&price, "price", "p", 0.0, "Purchase price in tokens")
	purchaseCmd.MarkFlagRequired("asset")
}
