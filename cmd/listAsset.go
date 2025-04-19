package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listAssetsCmd = &cobra.Command{
	Use:   "list-assets",
	Short: "List available assets on the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		db := storage.OpenDB("./data")
		defer db.CloseDB()

		blockchain := core.NewBlockchain(db)

		titleColor := color.New(color.FgCyan, color.Bold)
		headerColor := color.New(color.FgGreen, color.Bold)
		infoColor := color.New(color.FgWhite)
		hashColor := color.New(color.FgYellow)
		categoryColor := color.New(color.FgMagenta)

		// Print pretty header
		printCenteredTitle := func(title string) {
			width := 60
			padding := (width - len(title)) / 2
			titleColor.Println(strings.Repeat("=", width))
			titleColor.Println(strings.Repeat(" ", padding) + title)
			titleColor.Println(strings.Repeat("=", width))
		}

		printCenteredTitle("Available Digital Assets")

		// Track unique assets to avoid duplicates
		uniqueAssets := make(map[string]core.LicenseTransaction)

		// Iterate through blocks in reverse order to get the latest transactions
		for i := len(blockchain.Blocks) - 1; i >= 0; i-- {
			block := blockchain.Blocks[i]
			for _, tx := range block.Transaction {
				// Only add if we haven't seen this asset yet
				if _, exists := uniqueAssets[tx.AssetHash]; !exists {
					uniqueAssets[tx.AssetHash] = tx
				}
			}
		}

		if len(uniqueAssets) == 0 {
			fmt.Println("\n🔍 No assets found on the blockchain.")
			return
		}

		// Display assets
		count := 1
		for assetHash, tx := range uniqueAssets {
			fmt.Println()
			headerColor.Printf("Asset #%d\n", count)
			fmt.Println(strings.Repeat("-", 40))

			// Parse metadata if available
			var metadata map[string]string
			if err := json.Unmarshal([]byte(tx.Metadata), &metadata); err == nil {
				if title, ok := metadata["Title"]; ok {
					infoColor.Printf("📝 Title: ")
					fmt.Printf("%s\n", title)
				}
				if desc, ok := metadata["Description"]; ok {
					infoColor.Printf("📄 Description: ")
					fmt.Printf("%s\n", desc)
				}
				if category, ok := metadata["Category"]; ok {
					infoColor.Printf("🏷️ Category: ")
					categoryColor.Printf("%s\n", category)
				}
			}

			infoColor.Printf("🔑 License Type: ")
			fmt.Printf("%s\n", tx.License)

			infoColor.Printf("👤 Owner: ")
			fmt.Printf("%s\n", shortenKey(tx.Owner))

			infoColor.Printf("🆔 Asset ID: ")
			hashColor.Printf("%s\n", shortenHash(assetHash))

			// Display timestamp in human-readable format
			if tx.Timestamp > 0 {
				timeStr := time.Unix(tx.Timestamp, 0).Format("2006-01-02 15:04:05")
				infoColor.Printf("⏰ Created: ")
				fmt.Printf("%s\n", timeStr)
			}

			count++
		}
		fmt.Println()
	},
}

// Helper function to shorten hash for display
func shortenHash(hash string) string {
	if len(hash) <= 16 {
		return hash
	}
	return hash[:8] + "..." + hash[len(hash)-8:]
}

// Helper function to shorten public key for display
func shortenKey(key string) string {
	if len(key) <= 16 {
		return key
	}
	return key[:8] + "..." + key[len(key)-8:]
}

func init() {
	rootCmd.AddCommand(listAssetsCmd)
}
