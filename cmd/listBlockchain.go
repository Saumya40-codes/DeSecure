package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Saumya40-codes/DeSecure/core"
	storage "github.com/Saumya40-codes/DeSecure/pkg"
	"github.com/spf13/cobra"
)

var blockchainCmd = &cobra.Command{
	Use:   "blockchain",
	Short: "View current blockchain state",
	Run: func(cmd *cobra.Command, args []string) {
		db := storage.OpenDB("./data")
		defer db.CloseDB()

		blockchain := core.NewBlockchain(db)

		fmt.Printf("Blockchain Status - %d blocks\n\n", len(blockchain.Blocks))

		for i, block := range blockchain.Blocks {
			// Parse time
			t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", block.Timestamp)
			timeStr := block.Timestamp
			if err == nil {
				timeStr = t.Format("2006-01-02 15:04:05")
			}

			fmt.Printf("Block #%d\n", i)
			fmt.Printf("  Time: %s\n", timeStr)
			fmt.Printf("  Hash: %s\n", block.Hash)
			fmt.Printf("  Prev: %s\n", block.PrevHash)
			fmt.Printf("  Transactions: %d\n", len(block.Transaction))

			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose && len(block.Transaction) > 0 {
				fmt.Println("  Transaction Details:")
				for j, tx := range block.Transaction {
					fmt.Printf("    [%d] ID: %s\n", j, tx.TxID)
					fmt.Printf("        Owner: %s\n", tx.Owner)
					fmt.Printf("        Asset: %s\n", tx.AssetHash)
					fmt.Printf("        License: %s\n", tx.License)

					var metadata map[string]interface{}
					if err := json.Unmarshal([]byte(tx.Metadata), &metadata); err == nil {
						fmt.Println("        Metadata:")
						for k, v := range metadata {
							fmt.Printf("          %s: %v\n", k, v)
						}
					}
				}
			}

			// Add spacing between blocks
			if i < len(blockchain.Blocks)-1 {
				fmt.Println()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(blockchainCmd)
	blockchainCmd.Flags().BoolP("verbose", "v", false, "Show detailed transaction information")
}
