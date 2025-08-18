package holdmanager

import "testing"

func TestABI_HoldManager_HasExpectedMembers(t *testing.T) {
	abi, err := HoldmanagerMetaData.GetAbi()
	if err != nil {
		t.Fatalf("GetAbi: %v", err)
	}

	for _, m := range []string{
		"createHold", "executeHold", "releaseHold", "extendHold",
		"getHold", "getHoldStatus", "heldBalanceOf", "isExpired",
	} {
		if _, ok := abi.Methods[m]; !ok {
			t.Fatalf("expected method %q in HoldManager ABI", m)
		}
	}

	for _, e := range []string{
		"HoldCreated", "HoldExecuted", "HoldReleased", "HoldExtended", "HoldCanceled",
	} {
		if _, ok := abi.Events[e]; !ok {
			t.Fatalf("expected event %q in HoldManager ABI", e)
		}
	}
}
