package ledger

import "testing"

func TestTransferIsBalancedAndIdempotent(t *testing.T) {
	l := New()
	_, _ = l.CreateAccount("a", "USD", 1000)
	_, _ = l.CreateAccount("b", "USD", 0)
	first, err := l.Transfer("tx-1", "a", "b", 250)
	if err != nil {
		t.Fatal(err)
	}
	second, err := l.Transfer("tx-1", "a", "b", 250)
	if err != nil || first != second {
		t.Fatalf("idempotency failed: %v", err)
	}
	a, _ := l.Account("a")
	b, _ := l.Account("b")
	if a.BalanceCents != 750 || b.BalanceCents != 250 {
		t.Fatalf("unexpected balances: %+v %+v", a, b)
	}
	if err := l.Audit(); err != nil {
		t.Fatal(err)
	}
}
func TestRejectsOverdraft(t *testing.T) {
	l := New()
	_, _ = l.CreateAccount("a", "USD", 10)
	_, _ = l.CreateAccount("b", "USD", 0)
	if _, err := l.Transfer("tx", "a", "b", 11); err != ErrInsufficientFunds {
		t.Fatalf("expected overdraft error, got %v", err)
	}
}
