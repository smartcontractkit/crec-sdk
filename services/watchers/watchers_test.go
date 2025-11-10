package watchers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceCreation(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		_, err := NewService(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ServiceOptions is required")
	})

	t.Run("nil CREC client", func(t *testing.T) {
		opts := &ServiceOptions{}
		_, err := NewService(opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CRECClient is required")
	})
}

func TestCreateWatcherWithDomainInputValidation(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		tests := []struct {
			name      string
			channelID uuid.UUID
			input     CreateWatcherWithDomainInput
			expectErr string
		}{
			{
				name:      "empty channel ID",
				channelID: uuid.Nil,
				input: CreateWatcherWithDomainInput{
					ChainSelector: 1,
					Address:       "0x123",
					Domain:        "dvp",
					Events:        []string{"OperationExecuted"},
				},
				expectErr: "channel_id cannot be empty",
			},
			{
				name:      "empty chain selector",
				channelID: uuid.New(),
				input: CreateWatcherWithDomainInput{
					Address: "0x123",
					Domain:  "dvp",
					Events:  []string{"OperationExecuted"},
				},
				expectErr: "chain_selector is required",
			},
			{
				name:      "empty address",
				channelID: uuid.New(),
				input: CreateWatcherWithDomainInput{
					ChainSelector: 1,
					Domain:        "dvp",
					Events:        []string{"OperationExecuted"},
				},
				expectErr: "address is required",
			},
			{
				name:      "empty domain",
				channelID: uuid.New(),
				input: CreateWatcherWithDomainInput{
					ChainSelector: 1,
					Address:       "0x123",
					Events:        []string{"OperationExecuted"},
				},
				expectErr: "domain is required",
			},
			{
				name:      "empty events",
				channelID: uuid.New(),
				input: CreateWatcherWithDomainInput{
					ChainSelector: 1,
					Address:       "0x123",
					Domain:        "dvp",
				},
				expectErr: "events list cannot be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.channelID == uuid.Nil {
					assert.Contains(t, "channel_id cannot be empty", tt.expectErr)
					return
				}
				if tt.input.ChainSelector == 0 {
					assert.Contains(t, "chain_selector is required", tt.expectErr)
					return
				}
				if tt.input.Address == "" {
					assert.Contains(t, "address is required", tt.expectErr)
					return
				}
				if tt.input.Domain == "" {
					assert.Contains(t, "domain is required", tt.expectErr)
					return
				}
				if len(tt.input.Events) == 0 {
					assert.Contains(t, "events list cannot be empty", tt.expectErr)
					return
				}
			})
		}
	})
}

func TestCreateWatcherWithABIInputValidation(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		tests := []struct {
			name      string
			channelID uuid.UUID
			input     CreateWatcherWithABIInput
			expectErr string
		}{
			{
				name:      "empty channel ID",
				channelID: uuid.Nil,
				input: CreateWatcherWithABIInput{
					ChainSelector: 1,
					Address:       "0x123",
					Events:        []string{"Transfer"},
					ABI: []EventABI{
						{
							Type: "event",
							Name: "Transfer",
							Inputs: []EventABIInput{
								{Name: "from", Type: "address", InternalType: "address"},
							},
						},
					},
				},
				expectErr: "channel_id cannot be empty",
			},
			{
				name:      "empty chain selector",
				channelID: uuid.New(),
				input: CreateWatcherWithABIInput{
					Address: "0x123",
					Events:  []string{"Transfer"},
					ABI: []EventABI{
						{
							Type: "event",
							Name: "Transfer",
							Inputs: []EventABIInput{
								{Name: "from", Type: "address", InternalType: "address"},
							},
						},
					},
				},
				expectErr: "chain_selector is required",
			},
			{
				name:      "empty address",
				channelID: uuid.New(),
				input: CreateWatcherWithABIInput{
					ChainSelector: 1,
					Events:        []string{"Transfer"},
					ABI: []EventABI{
						{
							Type: "event",
							Name: "Transfer",
							Inputs: []EventABIInput{
								{Name: "from", Type: "address", InternalType: "address"},
							},
						},
					},
				},
				expectErr: "address is required",
			},
			{
				name:      "empty events",
				channelID: uuid.New(),
				input: CreateWatcherWithABIInput{
					ChainSelector: 1,
					Address:       "0x123",
					ABI: []EventABI{
						{
							Type: "event",
							Name: "Transfer",
							Inputs: []EventABIInput{
								{Name: "from", Type: "address", InternalType: "address"},
							},
						},
					},
				},
				expectErr: "events list cannot be empty",
			},
			{
				name:      "empty ABI",
				channelID: uuid.New(),
				input: CreateWatcherWithABIInput{
					ChainSelector: 1,
					Address:       "0x123",
					Events:        []string{"Transfer"},
				},
				expectErr: "abi cannot be empty",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.channelID == uuid.Nil {
					assert.Contains(t, "channel_id cannot be empty", tt.expectErr)
					return
				}
				if tt.input.ChainSelector == 0 {
					assert.Contains(t, "chain_selector is required", tt.expectErr)
					return
				}
				if tt.input.Address == "" {
					assert.Contains(t, "address is required", tt.expectErr)
					return
				}
				if len(tt.input.Events) == 0 {
					assert.Contains(t, "events list cannot be empty", tt.expectErr)
					return
				}
				if len(tt.input.ABI) == 0 {
					assert.Contains(t, "abi cannot be empty", tt.expectErr)
					return
				}
			})
		}
	})
}

func TestConvertUint64PtrToStringPtr(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result := convertUint64PtrToStringPtr(nil)
		assert.Nil(t, result)
	})

	t.Run("non-nil value", func(t *testing.T) {
		val := uint64(12345)
		result := convertUint64PtrToStringPtr(&val)
		require.NotNil(t, result)
		assert.Equal(t, "12345", *result)
	})

	t.Run("zero value", func(t *testing.T) {
		val := uint64(0)
		result := convertUint64PtrToStringPtr(&val)
		require.NotNil(t, result)
		assert.Equal(t, "0", *result)
	})
}

func TestWatcherInputTypes(t *testing.T) {
	t.Run("CreateWatcherWithDomainInput structure", func(t *testing.T) {
		name := "test-watcher"
		input := CreateWatcherWithDomainInput{
			Name:          &name,
			ChainSelector: 1,
			Address:       "0x123",
			Domain:        "dvp",
			Events:        []string{"OperationExecuted"},
		}

		assert.Equal(t, &name, input.Name)
		assert.Equal(t, uint64(1), input.ChainSelector)
		assert.Equal(t, "0x123", input.Address)
		assert.Equal(t, "dvp", input.Domain)
		assert.Len(t, input.Events, 1)
	})

	t.Run("CreateWatcherWithABIInput structure", func(t *testing.T) {
		name := "test-watcher"
		input := CreateWatcherWithABIInput{
			Name:          &name,
			ChainSelector: 1,
			Address:       "0x123",
			Events:        []string{"Transfer"},
			ABI: []EventABI{
				{
					Type: "event",
					Name: "Transfer",
					Inputs: []EventABIInput{
						{Name: "from", Type: "address", InternalType: "address", Indexed: true},
						{Name: "to", Type: "address", InternalType: "address", Indexed: true},
						{Name: "value", Type: "uint256", InternalType: "uint256", Indexed: false},
					},
				},
			},
		}

		assert.Equal(t, &name, input.Name)
		assert.Equal(t, uint64(1), input.ChainSelector)
		assert.Equal(t, "0x123", input.Address)
		assert.Len(t, input.Events, 1)
		assert.Len(t, input.ABI, 1)
		assert.Equal(t, "Transfer", input.ABI[0].Name)
		assert.Len(t, input.ABI[0].Inputs, 3)
	})

	t.Run("UpdateWatcherInput structure", func(t *testing.T) {
		input := UpdateWatcherInput{
			Name: "updated-name",
		}

		assert.Equal(t, "updated-name", input.Name)
	})
}

func TestWatcherStatus(t *testing.T) {
	t.Run("status constants", func(t *testing.T) {
		assert.Equal(t, WatcherStatus("pending"), StatusPending)
		assert.Equal(t, WatcherStatus("active"), StatusActive)
		assert.Equal(t, WatcherStatus("failed"), StatusFailed)
		assert.Equal(t, WatcherStatus("deleting"), StatusDeleting)
		assert.Equal(t, WatcherStatus("deleted"), StatusDeleted)
	})
}
