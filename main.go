package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

var blockchain *core.Blockchain
var db *storage.DB

func init() {
	db = storage.OpenDB("./badger_data")
	defer db.CloseDB()

	// TODO: Load blockchain state from DB
	blockchain = core.NewBlockchain()
}

func main() {
	if len(os.Args) >= 2 {
		if os.Args[1] == "list" {
			fmt.Println(blockchain)
			os.Exit(0)
		} else if os.Args[1] == "upload" {
			// TODO: Implement file upload logic
		} else {
			log.Fatal("Invalid arguments. Possible ones: list, upload")
		}
	}
}

