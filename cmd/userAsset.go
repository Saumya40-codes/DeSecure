package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Saumya40-codes/DeSecure/core"
	storage "github.com/Saumya40-codes/DeSecure/pkg"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var myAssetsCmd = &cobra.Command{
	Use:   "my-assets",
	Short: "List assets you own or have purchased licenses for",
	Run: func(cmd *cobra.Command, args []string) {
		db := storage.OpenDB("./data")
		defer db.CloseDB()

		blockchain := core.NewBlockchain(db)

		// Get user's public key
		_, pubKey := loadKeyPair()

		titleColor := color.New(color.FgCyan, color.Bold)
		headerColor := color.New(color.FgGreen, color.Bold)
		infoColor := color.New(color.FgWhite)
		hashColor := color.New(color.FgYellow)
		categoryColor := color.New(color.FgMagenta)
		roleColor := color.New(color.FgRed, color.Bold)

		// Print pretty header
		printCenteredTitle := func(title string) {
			width := 60
			padding := (width - len(title)) / 2
			titleColor.Println(strings.Repeat("=", width))
			titleColor.Println(strings.Repeat(" ", padding) + title)
			titleColor.Println(strings.Repeat("=", width))
		}

		printCenteredTitle("Your Digital Assets")

		// Track unique assets (owned or licensed)
		myAssets := make(map[string]core.LicenseTransaction)

		// Iterate through blocks to find relevant transactions
		for i := len(blockchain.Blocks) - 1; i >= 0; i-- {
			block := blockchain.Blocks[i]
			for _, tx := range block.Transaction {
				// If you're the owner or licensee and we haven't seen this asset yet
				if tx.Owner == pubKey || tx.Licensee == pubKey {
					if _, exists := myAssets[tx.AssetHash]; !exists {
						myAssets[tx.AssetHash] = tx
					}
				}
			}
		}

		if len(myAssets) == 0 {
			fmt.Println("\n🔍 You don't have any assets on the blockchain.")
			return
		}

		// Display assets
		count := 1
		for assetHash, tx := range myAssets {
			fmt.Println()
			headerColor.Printf("Asset #%d\n", count)
			fmt.Println(strings.Repeat("-", 40))

			// Parse metadata if available
			var metadata map[string]string
			if err := json.Unmarshal([]byte(tx.Metadata), &metadata); err == nil {
				if title, ok := metadata["Title"]; ok {
					infoColor.Printf(" Title: ")
					fmt.Printf("%s\n", title)
				}
				if desc, ok := metadata["Description"]; ok {
					infoColor.Printf(" Description: ")
					fmt.Printf("%s\n", desc)
				}
				if category, ok := metadata["Category"]; ok {
					infoColor.Printf(" Category: ")
					categoryColor.Printf("%s\n", category)
				}
			}

			infoColor.Printf(" License Type: ")
			fmt.Printf("%s\n", tx.License)

			// Show your role (owner or licensee)
			infoColor.Printf(" Your Role: ")
			if tx.Owner == pubKey {
				roleColor.Printf("Owner\n")
			} else {
				roleColor.Printf("Licensee\n")
				infoColor.Printf(" Owner: ")
				fmt.Printf("%s\n", shortenKey(tx.Owner))
			}

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

func init() {
	rootCmd.AddCommand(myAssetsCmd)
}
