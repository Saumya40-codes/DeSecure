package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
)

var (
	peerNodes    = []string{"http://node1:8080", "http://node2:8080"}
	networkMutex = sync.Mutex{}
)

// There can be many transaction in a block?? currently there is just one transac which might go into a block
// TODO: Check how/why there can be multiple of them
func BroadcastTransactions(tx LicenseTransaction) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	marshalContent, err := json.Marshal(tx)
	if err != nil {
		panic(err) // TODO: add loging
	}

	for _, node := range peerNodes {
		http.Post(node+"/receive_transaction", "application/json", bytes.NewBuffer(marshalContent))
	}
}
