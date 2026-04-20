package bundle_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/crec-sdk/extension/bundle"
)

func testBundle() *bundle.Bundle {
	return &bundle.Bundle{
		Service:        "dta",
		WasmBinary:     []byte("fake-wasm"),
		ConfigTemplate: []byte("fake-config"),
		Contracts: []bundle.Contract{
			{Name: "DTARequestManagement", ABI: `[{"type":"event","name":"Sub"}]`},
			{Name: "DTARequestSettlement", ABI: `[{"type":"event","name":"Settle"}]`},
		},
		Events: []bundle.Event{
			{
				Name:            "SubscriptionRequested",
				TriggerContract: "DTARequestManagement",
				Description:     "Investor subscription request",
				ParamsSchema:    json.RawMessage(`{"type":"object"}`),
			},
			{
				Name:            "DTASettlementOpened",
				TriggerContract: "DTARequestSettlement",
				Description:     "DTA settlement initiated",
				ParamsSchema:    json.RawMessage(`{"type":"object"}`),
			},
		},
	}
}

func TestBundle_FindTriggerContract(t *testing.T) {
	tests := []struct {
		name         string
		eventName    string
		wantContract string
		wantNil      bool
	}{
		{
			name:         "finds DTARequestManagement for SubscriptionRequested",
			eventName:    "SubscriptionRequested",
			wantContract: "DTARequestManagement",
		},
		{
			name:         "finds DTARequestSettlement for DTASettlementOpened",
			eventName:    "DTASettlementOpened",
			wantContract: "DTARequestSettlement",
		},
		{
			name:      "returns nil for unknown event",
			eventName: "NonExistentEvent",
			wantNil:   true,
		},
		{
			name:      "returns nil for empty event name",
			eventName: "",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBundle()
			contract := b.FindTriggerContract(tt.eventName)

			if tt.wantNil {
				assert.Nil(t, contract)
			} else {
				require.NotNil(t, contract)
				assert.Equal(t, tt.wantContract, contract.Name)
			}
		})
	}
}

func TestBundle_FindTriggerContract_MissingContract(t *testing.T) {
	b := &bundle.Bundle{
		Service: "test",
		Events: []bundle.Event{
			{
				Name:            "Orphan",
				TriggerContract: "NonExistentContract",
			},
		},
		Contracts: []bundle.Contract{
			{Name: "SomeOtherContract", ABI: "[]"},
		},
	}

	contract := b.FindTriggerContract("Orphan")
	assert.Nil(t, contract, "should return nil when TriggerContract name doesn't match any Contract")
}

func TestBundle_HasEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		want      bool
	}{
		{
			name:      "returns true for existing event",
			eventName: "SubscriptionRequested",
			want:      true,
		},
		{
			name:      "returns true for second event",
			eventName: "DTASettlementOpened",
			want:      true,
		},
		{
			name:      "returns false for non-existent event",
			eventName: "DoesNotExist",
			want:      false,
		},
		{
			name:      "returns false for empty event name",
			eventName: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := testBundle()
			assert.Equal(t, tt.want, b.HasEvent(tt.eventName))
		})
	}
}

func TestBundle_HasEvent_EmptyBundle(t *testing.T) {
	b := &bundle.Bundle{Service: "empty"}
	assert.False(t, b.HasEvent("anything"))
}

func TestBundle_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bundle  *bundle.Bundle
		wantErr string
	}{
		{
			name:    "valid bundle passes",
			bundle:  testBundle(),
			wantErr: "",
		},
		{
			name: "empty WasmBinary",
			bundle: &bundle.Bundle{
				Service:        "test",
				ConfigTemplate: []byte("tmpl"),
			},
			wantErr: "WasmBinary is empty",
		},
		{
			name: "nil ConfigTemplate passes validation",
			bundle: &bundle.Bundle{
				Service:    "test",
				WasmBinary: []byte("wasm"),
			},
			wantErr: "",
		},
		{
			name: "duplicate contract name",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Contracts: []bundle.Contract{
					{Name: "Foo", ABI: "[]"},
					{Name: "Foo", ABI: "[]"},
				},
			},
			wantErr: "duplicate contract name",
		},
		{
			name: "duplicate event name",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Contracts: []bundle.Contract{
					{Name: "Foo", ABI: "[]"},
				},
				Events: []bundle.Event{
					{Name: "Evt", TriggerContract: "Foo"},
					{Name: "Evt", TriggerContract: "Foo"},
				},
			},
			wantErr: "duplicate event name",
		},
		{
			name: "orphaned TriggerContract",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Contracts: []bundle.Contract{
					{Name: "Real", ABI: "[]"},
				},
				Events: []bundle.Event{
					{Name: "Evt", TriggerContract: "DoesNotExist"},
				},
			},
			wantErr: "does not match any contract",
		},
		{
			name: "invalid ABI JSON",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Contracts: []bundle.Contract{
					{Name: "Bad", ABI: "{not json"},
				},
			},
			wantErr: "not valid JSON",
		},
		{
			name: "empty contract name",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Contracts: []bundle.Contract{
					{Name: "", ABI: "[]"},
				},
			},
			wantErr: "Name is empty",
		},
		{
			name: "empty event name",
			bundle: &bundle.Bundle{
				Service:        "test",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
				Events: []bundle.Event{
					{Name: ""},
				},
			},
			wantErr: "Name is empty",
		},
		{
			name: "generic bundle with no contracts or events passes",
			bundle: &bundle.Bundle{
				Service:        "",
				WasmBinary:     []byte("wasm"),
				ConfigTemplate: []byte("tmpl"),
			},
			wantErr: "",
		},
		{
			name: "multiple errors reported",
			bundle: &bundle.Bundle{
				Service: "test",
			},
			wantErr: "WasmBinary is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bundle.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBundle_DefaultConfigTemplate(t *testing.T) {
	tmpl := bundle.DefaultConfigTemplate()
	require.NotEmpty(t, tmpl, "default config template should be embedded")
	assert.Contains(t, string(tmpl), "detectEventTriggerConfig")
	assert.Contains(t, string(tmpl), "contractName")
	assert.Contains(t, string(tmpl), "contractReaderConfig")
	assert.Contains(t, string(tmpl), `confidenceLevel: "{{ $config.confidenceLevel }}"`)
}

func TestBundle_ResolveConfigTemplate(t *testing.T) {
	tests := []struct {
		name           string
		configTemplate []byte
		wantDefault    bool
	}{
		{
			name:           "nil ConfigTemplate returns default",
			configTemplate: nil,
			wantDefault:    true,
		},
		{
			name:           "empty ConfigTemplate returns default",
			configTemplate: []byte{},
			wantDefault:    true,
		},
		{
			name:           "custom ConfigTemplate is returned",
			configTemplate: []byte("custom-template-content"),
			wantDefault:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bundle.Bundle{
				ConfigTemplate: tt.configTemplate,
			}
			result := bundle.ResolveConfigTemplate(b)
			if tt.wantDefault {
				assert.Equal(t, bundle.DefaultConfigTemplate(), result)
			} else {
				assert.Equal(t, tt.configTemplate, result)
			}
		})
	}
}
