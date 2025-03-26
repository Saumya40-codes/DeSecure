package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
)

var blockchain *core.Blockchain

func init() {
	// TODO: add persistance using some kv db, currently on every run, an empty blockchain is getting created
	blockchain = core.NewBlockchain()
}
func main() {
	if len(os.Args) >= 2 {
		fmt.Println(os.Args)
		// Check if its 'list' command
		if os.Args[1] == "list" {
			fmt.Println(blockchain)
			os.Exit(0)
		} else {
			// unintended command
			log.Fatal("Invalid arguments passed. Possible ones for now: list")
		}
	}
}

