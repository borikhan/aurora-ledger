package ledger

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrDuplicateAccount  = errors.New("account already exists")
	ErrAccountNotFound   = errors.New("account not found")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type Entry struct {
	TransferID string    `json:"transferId"`
	AccountID  string    `json:"accountId"`
	DeltaCents int64     `json:"deltaCents"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Account struct {
	ID           string `json:"id"`
	Currency     string `json:"currency"`
	BalanceCents int64  `json:"balanceCents"`
}

type Transfer struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	AmountCents int64     `json:"amountCents"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Ledger struct {
	mu        sync.RWMutex
	accounts  map[string]*Account
	transfers map[string]Transfer
	entries   []Entry
}

func New() *Ledger {
	return &Ledger{accounts: map[string]*Account{}, transfers: map[string]Transfer{}}
}

func (l *Ledger) CreateAccount(id, currency string, openingBalance int64) (Account, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if id == "" || currency == "" || openingBalance < 0 {
		return Account{}, errors.New("invalid account")
	}
	if _, ok := l.accounts[id]; ok {
		return Account{}, ErrDuplicateAccount
	}
	a := &Account{ID: id, Currency: currency, BalanceCents: openingBalance}
	l.accounts[id] = a
	return *a, nil
}

func (l *Ledger) Transfer(id, from, to string, amount int64) (Transfer, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if existing, ok := l.transfers[id]; ok {
		return existing, nil
	}
	if amount <= 0 || from == to {
		return Transfer{}, ErrInvalidAmount
	}
	source, sourceOK := l.accounts[from]
	target, targetOK := l.accounts[to]
	if !sourceOK || !targetOK {
		return Transfer{}, ErrAccountNotFound
	}
	if source.Currency != target.Currency {
		return Transfer{}, errors.New("currency mismatch")
	}
	if source.BalanceCents < amount {
		return Transfer{}, ErrInsufficientFunds
	}
	now := time.Now().UTC()
	t := Transfer{ID: id, From: from, To: to, AmountCents: amount, CreatedAt: now}
	source.BalanceCents -= amount
	target.BalanceCents += amount
	l.entries = append(l.entries,
		Entry{TransferID: id, AccountID: from, DeltaCents: -amount, CreatedAt: now},
		Entry{TransferID: id, AccountID: to, DeltaCents: amount, CreatedAt: now},
	)
	l.transfers[id] = t
	return t, nil
}

func (l *Ledger) Account(id string) (Account, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	a, ok := l.accounts[id]
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return *a, nil
}

func (l *Ledger) Audit() error {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var sum int64
	for _, entry := range l.entries {
		sum += entry.DeltaCents
	}
	if sum != 0 {
		return fmt.Errorf("unbalanced ledger: %d", sum)
	}
	return nil
}
