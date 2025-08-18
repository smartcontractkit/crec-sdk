package erc20

import "testing"

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
