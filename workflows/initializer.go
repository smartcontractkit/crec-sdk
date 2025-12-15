package workflows

import (
	"fmt"
	"strconv"

	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	"github.com/smartcontractkit/cre-sdk-go/cre"
)

// LogHandler is the function signature implemented by each event-listener workflow's
// per-project handler (e.g. OnLog, OnCoordinatorLog). It processes an EVM log
// and returns a base64-encoded verifiable event (or empty string) and an error.
type LogHandler func(*Config, cre.Runtime, *evm.Log) (string, error)

// InitEventListenerWorkflow wires the standard EVM Log trigger for event-listener
// workflows and attaches the provided handler. It resolves the event signature
// from the ABI and uses cfg.ChainSelector (required in the config).
func InitEventListenerWorkflow(cfg *Config, handler LogHandler) (cre.Workflow[*Config], error) {
	abiJSON, err := GetContractABI(cfg, cfg.DetectEventTriggerConfig.ContractName)
	if err != nil {
		return nil, err
	}
	ev := MustEvent(abiJSON, cfg.DetectEventTriggerConfig.ContractEventName)
	eventSigHash := ev.ID.Bytes()

	filter := NewEVMLogFilter(cfg.DetectEventTriggerConfig.ContractAddress, eventSigHash)
	// Convert chainSelector string to uint64 for EVM client
	chainSelector, err := strconv.ParseUint(cfg.ChainSelector, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chain selector: %w", err)
	}

	return cre.Workflow[*Config]{
		cre.Handler(
			evm.LogTrigger(chainSelector, filter),
			handler,
		),
	}, nil
}
