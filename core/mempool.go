package core

import "sync"

type Mempool struct {
	mu           sync.Mutex
	transactions []LicenseTransaction
}

func NewMempool() *Mempool {
	return &Mempool{
		mu:           sync.Mutex{},
		transactions: []LicenseTransaction{},
	}
}

func (m *Mempool) AddTransaction(tx LicenseTransaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transactions = append(m.transactions, tx)
}

func (m *Mempool) GetTransactions() []LicenseTransaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.transactions
}

func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.transactions = []LicenseTransaction{}
}
