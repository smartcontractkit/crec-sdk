package bundle

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed config.tmpl
var defaultConfigTemplate []byte

// DefaultConfigTemplate returns the shared gomplate template for generating
// config.yaml. This covers the standard watcher config used extensions.
func DefaultConfigTemplate() []byte {
	return defaultConfigTemplate
}

// ResolveConfigTemplate returns the bundle's custom template if set,
// otherwise the shared default. Extensions with custom config fields
// can override the default by setting Bundle.ConfigTemplate.
func ResolveConfigTemplate(b *Bundle) []byte {
	if len(b.ConfigTemplate) > 0 {
		return b.ConfigTemplate
	}
	return defaultConfigTemplate
}

// Bundle is the deployment artifact for an extension watcher workflow.
// Extensions produce these; the CREC backend consumes them.
type Bundle struct {
	// Service identifies the extension (e.g. "dta", "dvp", "test_consumer").
	// An empty string means this is the generic/fallback event-listener.
	Service string

	// WasmBinary is the pre-compiled WASM binary, populated via //go:embed.
	WasmBinary []byte

	// ConfigTemplate is an optional gomplate template for generating config.yaml.
	// When nil, ResolveConfigTemplate falls back to the shared default template.
	// Set this only when an extension needs custom config fields beyond the
	// standard event-listener structure.
	ConfigTemplate []byte

	// Contracts lists the on-chain contracts this extension interacts with.
	// Nil for the generic bundle (user provides ABI at deploy time).
	Contracts []Contract

	// Events lists the subscribable events this extension supports.
	// Nil for the generic bundle (user provides events at deploy time).
	Events []Event
}

// Contract describes a contract the watcher needs access to.
type Contract struct {
	// Name is the logical contract identifier (e.g. "DTARequestManagement").
	Name string

	// ABI is the full ABI JSON string (or subset of event signatures) for the contract.
	ABI string
}

// Event describes a subscribable event.
type Event struct {
	// Name is the Solidity event name (e.g. "SubscriptionRequested").
	Name string

	// TriggerContract is the Contract.Name that emits this event.
	TriggerContract string

	// Description is a human-readable summary of the event.
	Description string

	// ParamsSchema is the JSON Schema for decoded Solidity event parameters (chain_event.params).
	ParamsSchema json.RawMessage

	// DataSchema is the JSON Schema for enrichment/reference data, or nil if no enrichment.
	DataSchema json.RawMessage
}

// Validate checks the bundle for configuration errors.
// It verifies that required fields are set, that contract and event names
// are unique, that ABI strings are valid JSON, and that every event's
// TriggerContract references an existing contract.
func (b *Bundle) Validate() error {
	var errs []string

	if len(b.WasmBinary) == 0 {
		errs = append(errs, "WasmBinary is empty")
	}
	contractNames := make(map[string]bool, len(b.Contracts))
	for i, c := range b.Contracts {
		if c.Name == "" {
			errs = append(errs, fmt.Sprintf("Contracts[%d].Name is empty", i))
			continue
		}
		if contractNames[c.Name] {
			errs = append(errs, fmt.Sprintf("duplicate contract name %q", c.Name))
		}
		contractNames[c.Name] = true
		if c.ABI != "" && !json.Valid([]byte(c.ABI)) {
			errs = append(errs, fmt.Sprintf("Contracts[%q].ABI is not valid JSON", c.Name))
		}
	}

	eventNames := make(map[string]bool, len(b.Events))
	for i, evt := range b.Events {
		if evt.Name == "" {
			errs = append(errs, fmt.Sprintf("Events[%d].Name is empty", i))
			continue
		}
		if eventNames[evt.Name] {
			errs = append(errs, fmt.Sprintf("duplicate event name %q", evt.Name))
		}
		eventNames[evt.Name] = true

		if evt.TriggerContract != "" && !contractNames[evt.TriggerContract] {
			errs = append(errs, fmt.Sprintf("Events[%q].TriggerContract %q does not match any contract", evt.Name, evt.TriggerContract))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("bundle validation failed: %s", strings.Join(errs, "; "))
	}
	return nil
}

// FindTriggerContract returns the Contract that emits the given event,
// or nil if not found.
func (b *Bundle) FindTriggerContract(eventName string) *Contract {
	var triggerName string
	for _, evt := range b.Events {
		if evt.Name == eventName {
			triggerName = evt.TriggerContract
			break
		}
	}
	if triggerName == "" {
		return nil
	}
	for i := range b.Contracts {
		if b.Contracts[i].Name == triggerName {
			return &b.Contracts[i]
		}
	}
	return nil
}

// HasEvent reports whether the bundle contains the named event.
func (b *Bundle) HasEvent(eventName string) bool {
	for _, evt := range b.Events {
		if evt.Name == eventName {
			return true
		}
	}
	return false
}
