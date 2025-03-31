package main

import (
	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/cmd"
	"github.com/Saumya40-codes/Hopefully_a_blockchain_project/core"
	storage "github.com/Saumya40-codes/Hopefully_a_blockchain_project/pkg"
)

var (
	blockchain *core.Blockchain
	db         *storage.DB
)

func init() {
	db = storage.OpenDB("./data")
	defer db.CloseDB()

	// TODO: Load blockchain state from DB
	blockchain = core.NewBlockchain()
}

func main() {
	cmd.Execute()
	select {}
}
