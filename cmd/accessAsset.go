package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
	"github.com/spf13/cobra"
)

var accessCmd = &cobra.Command{
	Use:   "access",
	Short: "Access content you've purchased or own",
	Run: func(cmd *cobra.Command, args []string) {
		if assetID == "" {
			fmt.Println("❌ Please specify an asset ID with -a flag")
			return
		}

		_, pubKey := loadKeyPair()

		db := storage.OpenDB("./data")
		defer db.CloseDB()

		blockchain := core.NewBlockchain(db)

		// Check if user has valid license
		if !core.HasValidLicense(assetID, blockchain, pubKey) {
			fmt.Println("❌ You don't have a valid license for this asset")
			return
		}

		gatewayURL := fmt.Sprintf("https://ipfs.io/ipfs/%s", assetID)

		fmt.Println("✅ License verified for asset:", shortenHash(assetID))
		fmt.Println("🌐 Accessing content from IPFS...")

		openBrowser(gatewayURL)

		fmt.Println("If your browser doesn't open automatically, access the content at:")
		fmt.Println(gatewayURL)
	},
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}

func init() {
	rootCmd.AddCommand(accessCmd)
	accessCmd.Flags().StringVarP(&assetID, "asset", "a", "", "Asset ID/hash to access")
	accessCmd.MarkFlagRequired("asset")
}
