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

// Get a transaction by ID
func (m *Mempool) GetTransactionByID(txID string) *LicenseTransaction {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, tx := range m.transactions {
		if tx.TxID == txID {
			return &m.transactions[i]
		}
	}
	return nil
}

// Remove a transaction by ID
func (m *Mempool) RemoveTransaction(txID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, tx := range m.transactions {
		if tx.TxID == txID {
			// Remove the transaction from the slice
			m.transactions = append(m.transactions[:i], m.transactions[i+1:]...)
			return
		}
	}
}
