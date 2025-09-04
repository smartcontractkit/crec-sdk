package dta

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toB32(hexStr string) [32]byte {
	var out [32]byte
	raw := hexStr
	if len(raw) >= 2 && (raw[0:2] == "0x" || raw[0:2] == "0X") {
		raw = raw[2:]
	}
	if len(raw)%2 != 0 {
		raw = "0" + raw
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		panic(err)
	}
	if len(b) != 32 {
		panic("invalid length for bytes32")
	}
	copy(out[:], b)
	return out
}

func TestVerifiableEvent_ParseAttributes(t *testing.T) {
	tests := []struct {
		name          string
		event         VerifiableEvent
		expectedError string
		expectedValue any
	}{
		{
			name:          "noAttributes",
			event:         VerifiableEvent{},
			expectedError: "no attributes",
		},
		{
			name: "unknownEventType",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type": {Value: "unknown_event_type"},
						},
					},
				},
			},
			expectedError: "unknown event type: ",
			expectedValue: nil,
		},
		{
			name: "DistributorRegistered",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "DistributorRegistered"},
							"distributor_addr": {Value: "0x0000000000000000000000000000000000000001"},
						}}}},
			expectedValue: DistributorRegistered{DistributorAddr: common.HexToAddress("0x0000000000000000000000000000000000000001")},
		},
		{
			name: "DistributorRequestCanceled",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "DistributorRequestCanceled"},
							"fund_token_id":    {Value: "0x0000000000000000000000000000000000000000000000000000000000000000"},
							"distributor_addr": {Value: "0x0000000000000000000000000000000000000002"},
							"request_id":       {Value: "0x0000000000000000000000000000000000000000000000000000000000000001"},
						}}}},
			expectedValue: DistributorRequestCanceled{
				FundTokenId:     toB32("0x0000000000000000000000000000000000000000000000000000000000000000"),
				DistributorAddr: common.HexToAddress("0x0000000000000000000000000000000000000002"),
				RequestId:       toB32("0x0000000000000000000000000000000000000000000000000000000000000001"),
			},
		},
		{
			name: "DistributorRequestProcessed",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type": {Value: "DistributorRequestProcessed"},
							"request_id": {Value: "0x00000000000000000000000000000000000000000000000000000000000000aa"},
							"shares":     {Value: "1000"},
							"status":     {Value: "1"},
						}}}},
			expectedValue: DistributorRequestProcessed{
				RequestId: toB32("0x00000000000000000000000000000000000000000000000000000000000000aa"),
				Shares:    big.NewInt(1000),
				Status:    1,
			},
		},
		{
			name: "DistributorRequestProcessing",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "DistributorRequestProcessing"},
							"fund_token_id":    {Value: "0x00000000000000000000000000000000000000000000000000000000000000bb"},
							"distributor_addr": {Value: "0x0000000000000000000000000000000000000003"},
							"request_id":       {Value: "0x00000000000000000000000000000000000000000000000000000000000000cc"},
							"shares":           {Value: "2000"},
							"amount":           {Value: "3000"},
						}}}},
			expectedValue: DistributorRequestProcessing{
				FundTokenId:     toB32("0x00000000000000000000000000000000000000000000000000000000000000bb"),
				DistributorAddr: common.HexToAddress("0x0000000000000000000000000000000000000003"),
				RequestId:       toB32("0x00000000000000000000000000000000000000000000000000000000000000cc"),
				Shares:          big.NewInt(2000),
				Amount:          big.NewInt(3000),
			},
		},
		{
			name: "FundAdminRegistered",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":      {Value: "FundAdminRegistered"},
							"fund_admin_addr": {Value: "0x0000000000000000000000000000000000000004"},
						}}}},
			expectedValue: FundAdminRegistered{FundAdminAddr: common.HexToAddress("0x0000000000000000000000000000000000000004")},
		},
		{
			name: "FundTokenAllowlistUpdated",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "FundTokenAllowlistUpdated"},
							"fund_admin_addr":  {Value: "0x0000000000000000000000000000000000000005"},
							"fund_token_id":    {Value: "0x00000000000000000000000000000000000000000000000000000000000000dd"},
							"distributor_addr": {Value: "0x0000000000000000000000000000000000000006"},
							"allowed":          {Value: "true"},
						}}}},
			expectedValue: FundTokenAllowlistUpdated{
				FundAdminAddr:   common.HexToAddress("0x0000000000000000000000000000000000000005"),
				FundTokenId:     toB32("0x00000000000000000000000000000000000000000000000000000000000000dd"),
				DistributorAddr: common.HexToAddress("0x0000000000000000000000000000000000000006"),
				Allowed:         true,
			},
		},
		{
			name: "FundTokenRegistered",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":           {Value: "FundTokenRegistered"},
							"fund_admin_addr":      {Value: "0x0000000000000000000000000000000000000007"},
							"fund_token_id":        {Value: "0x00000000000000000000000000000000000000000000000000000000000000ee"},
							"fund_token_addr":      {Value: "0x0000000000000000000000000000000000000008"},
							"nav_addr":             {Value: "0x0000000000000000000000000000000000000009"},
							"token_chain_selector": {Value: "12345"},
						}}}},
			expectedValue: FundTokenRegistered{
				FundAdminAddr:      common.HexToAddress("0x0000000000000000000000000000000000000007"),
				FundTokenId:        toB32("0x00000000000000000000000000000000000000000000000000000000000000ee"),
				FundTokenAddr:      common.HexToAddress("0x0000000000000000000000000000000000000008"),
				NavAddr:            common.HexToAddress("0x0000000000000000000000000000000000000009"),
				TokenChainSelector: 12345,
			},
		},
		{
			name: "SubscriptionRequested",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "SubscriptionRequested"},
							"fund_token_id":    {Value: "0x00000000000000000000000000000000000000000000000000000000000000ff"},
							"distributor_addr": {Value: "0x000000000000000000000000000000000000000a"},
							"request_id":       {Value: "0x0000000000000000000000000000000000000000000000000000000000000100"},
							"amount":           {Value: "5500"},
							"created_at":       {Value: "1690000000"},
						}}}},
			expectedValue: SubscriptionRequested{
				FundTokenId:     toB32("0x00000000000000000000000000000000000000000000000000000000000000ff"),
				DistributorAddr: common.HexToAddress("0x000000000000000000000000000000000000000a"),
				RequestId:       toB32("0x0000000000000000000000000000000000000000000000000000000000000100"),
				Amount:          big.NewInt(5500),
				CreatedAt:       1690000000,
			},
		},
		{
			name: "RedemptionRequested",
			event: VerifiableEvent{
				Metadata: Metadata{
					WorkflowEvent: WorkflowEvent{
						Attributes: map[string]Attribute{
							"event_type":       {Value: "RedemptionRequested"},
							"fund_token_id":    {Value: "0x0000000000000000000000000000000000000000000000000000000000000200"},
							"distributor_addr": {Value: "0x000000000000000000000000000000000000000b"},
							"request_id":       {Value: "0x0000000000000000000000000000000000000000000000000000000000000201"},
							"shares":           {Value: "777"},
							"created_at":       {Value: "1690000001"},
						}}}},
			expectedValue: RedemptionRequested{
				FundTokenId:     toB32("0x0000000000000000000000000000000000000000000000000000000000000200"),
				DistributorAddr: common.HexToAddress("0x000000000000000000000000000000000000000b"),
				RequestId:       toB32("0x0000000000000000000000000000000000000000000000000000000000000201"),
				Shares:          big.NewInt(777),
				CreatedAt:       1690000001,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.event.ParseAttributes()
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedValue, result)
		})
	}
}
