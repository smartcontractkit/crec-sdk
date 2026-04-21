package erc20

import (
	"testing"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestABI_ERC20_HasExpectedMembers(t *testing.T) {
	abi, err := Erc20MetaData.GetAbi()
	if err != nil {
		t.Fatalf("GetAbi: %v", err)
	}

	// Core methods
	for _, m := range []string{"transfer", "approve", "balanceOf", "totalSupply", "transferFrom"} {
		if _, ok := abi.Methods[m]; !ok {
			t.Fatalf("expected method %q in ERC20 ABI", m)
		}
	}

	// Core events
	for _, e := range []string{"Transfer", "Approval"} {
		if _, ok := abi.Events[e]; !ok {
			t.Fatalf("expected event %q in ERC20 ABI", e)
		}
	}
}

// MockSubscription implements event.Subscription
type mockSubscription struct {
	errChan chan error
	unsub   func()
}

func (m *mockSubscription) Unsubscribe() {
	if m.unsub != nil {
		m.unsub()
	}
}

func (m *mockSubscription) Err() <-chan error {
	return m.errChan
}

func TestErc20ApprovalIterator_Close_DrainsLogs(t *testing.T) {
	// Create an iterator with a buffered channel filled with some mock logs
	logsChan := make(chan types.Log, 5)
	
	// Add 3 dummy logs
	for i := 0; i < 3; i++ {
		logsChan <- types.Log{}
	}
	
	unsubCalled := false
	sub := &mockSubscription{
		errChan: make(chan error),
		unsub: func() {
			unsubCalled = true
		},
	}
	
	it := &Erc20ApprovalIterator{
		logs: logsChan,
		sub:  sub,
	}
	
	// Call Close to trigger the channel draining
	err := it.Close()
	if err != nil {
		t.Fatalf("expected no error from Close, got %v", err)
	}
	
	if !unsubCalled {
		t.Fatal("expected Unsubscribe to be called")
	}
	
	// The channel should now be empty
	select {
	case <-logsChan:
		t.Fatal("expected logs channel to be empty after Close drained it")
	default:
		// Success
	}
}

func TestErc20TransferIterator_Close_DrainsLogs(t *testing.T) {
	// Create an iterator with a buffered channel filled with some mock logs
	logsChan := make(chan types.Log, 5)
	
	// Add 3 dummy logs
	for i := 0; i < 3; i++ {
		logsChan <- types.Log{}
	}
	
	unsubCalled := false
	sub := &mockSubscription{
		errChan: make(chan error),
		unsub: func() {
			unsubCalled = true
		},
	}
	
	it := &Erc20TransferIterator{
		logs: logsChan,
		sub:  sub,
	}
	
	// Call Close to trigger the channel draining
	err := it.Close()
	if err != nil {
		t.Fatalf("expected no error from Close, got %v", err)
	}
	
	if !unsubCalled {
		t.Fatal("expected Unsubscribe to be called")
	}
	
	// The channel should now be empty
	select {
	case <-logsChan:
		t.Fatal("expected logs channel to be empty after Close drained it")
	default:
		// Success
	}
}
