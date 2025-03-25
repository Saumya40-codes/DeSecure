package core

import "sync"

var (
	peerNodes   = []string{"http://node1:8080", "http://node2:8080"}
	networkLock = sync.Mutex{}
)

func BroadcastTransactions(tx LicenseTransaction) {

}
