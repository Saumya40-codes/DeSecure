package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drmcli",
	Short: "DRMCLI is a command-line tool for decentralized DRM",
	Long:  `A command-line tool for managing decentralized digital rights, including file uploads and access verification.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
