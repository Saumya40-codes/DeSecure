package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

var (
	blockchain *core.Blockchain
	db         *storage.DB
)

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
			filePath := flag.String("path", "", "Path of file to be uploaded")

			var permissions string
			flag.StringVar(&permissions, "perm", "read", "Permission to be attached with the file. Default [read]. Comma seperated value of permissions [read,write,download]")

			flag.Parse()

			if *filePath == "" {
				log.Fatal("Please provide a file path using -path flag")
			}
		} else {
			log.Fatal("Invalid arguments. Possible ones: list, upload")
		}
	}
}
