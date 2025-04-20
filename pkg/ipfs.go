package storage

import (
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

func UploadtoIPFS(filePath string) (string, error) {
	sh := shell.NewShell("localhost:5001")

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return "", err
	}

	cid, err := sh.Add(file)
	if err != nil {
		return "", err
	}

	return cid, nil
}
